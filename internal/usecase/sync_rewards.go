package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/model"
)

// SyncRewards handles business logic for syncing rewards.
type SyncRewards struct {
	batchSize      int
	dbAdapter      database.Adapter
	logger         *logrus.Entry
	tzktApiAdapter tzktapi.Adapter
}

// NewSyncRewardsFunc creates a new instance of SyncRewards and returns a SyncFunc.
func NewSyncRewardsFunc(tzktAdapter tzktapi.Adapter, dbAdapter database.Adapter, metricsClient metrics.Adapter, logger *logrus.Entry) model.SyncFunc {
	uc := &SyncRewards{
		batchSize:      1000,
		dbAdapter:      dbAdapter,
		logger:         logger.WithField("usecase", "sync_rewards"),
		tzktApiAdapter: tzktAdapter,
	}
	return uc.withMonitorer(uc.Sync, metricsClient)
}

// Sync syncs rewards from the TzKT API to the database.
func (uc *SyncRewards) Sync(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
	}

	// Get the current cycle from TzKT API
	currentCycle, err := uc.tzktApiAdapter.GetCurrentCycle(ctx)
	if err != nil {
		return fmt.Errorf("error fetching current cycle: %w", err)
	}

	// Get the last synced cycle from the database
	lastSyncedCycle, err := uc.dbAdapter.GetLastSyncedRewardCycle(ctx)
	if err != nil {
		uc.logger.Warnf("Error getting last synced rewards cycle: %v, assuming 0", err)
		lastSyncedCycle = 0
	}

	startCycle := lastSyncedCycle + 1
	if startCycle > currentCycle {
		uc.logger.Info("Rewards are already up to date")
		return nil
	}

	uc.logger.Infof("Syncing rewards from cycle %d to cycle %d", startCycle, currentCycle)

	// Fetch active delegators
	delegators, err := uc.dbAdapter.GetActiveDelegators(ctx)
	if err != nil {
		return fmt.Errorf("error fetching active delegators: %w", err)
	}

	if len(delegators) == 0 {
		uc.logger.Info("No active delegators found")
		return nil
	}

	uc.logger.Infof("Found %d active delegators", len(delegators))

	// For each cycle that needs to be synced
	for cycle := startCycle; cycle <= currentCycle; cycle++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		uc.logger.Infof("Processing rewards for cycle %d", cycle)

		// For each delegator, fetch their rewards for this cycle
		var allRewards []model.Reward
		for _, delegator := range delegators {
			// Get the baker for this delegator during this cycle
			baker, err := uc.dbAdapter.GetBakerForDelegatorAtCycle(ctx, delegator, cycle)
			if err != nil {
				uc.logger.Warnf("Error getting baker for delegator %s at cycle %d: %v, skipping", delegator, cycle, err)
				continue
			}

			if baker == "" {
				uc.logger.Debugf("No baker found for delegator %s at cycle %d, skipping", delegator, cycle)
				continue
			}

			// Fetch rewards from TzKT
			rewards, err := uc.tzktApiAdapter.FetchRewardsForCycle(ctx, delegator, baker, cycle)
			if err != nil {
				uc.logger.Warnf("Error fetching rewards for delegator %s, baker %s, cycle %d: %v, skipping", delegator, baker, cycle, err)
				continue
			}

			if len(rewards) == 0 {
				uc.logger.Debugf("No rewards found for delegator %s, baker %s, cycle %d", delegator, baker, cycle)
				continue
			}

			allRewards = append(allRewards, rewards...)
		}

		if len(allRewards) > 0 {
			// Save rewards to database in batches
			if err := uc.saveRewardsBatch(ctx, allRewards, cycle); err != nil {
				return fmt.Errorf("error saving rewards for cycle %d: %w", cycle, err)
			}
			uc.logger.Infof("Saved %d rewards for cycle %d", len(allRewards), cycle)
		} else {
			uc.logger.Infof("No rewards found for cycle %d", cycle)
		}

		// Update the last synced cycle in the database
		if err := uc.dbAdapter.SaveLastSyncedRewardCycle(ctx, cycle); err != nil {
			return fmt.Errorf("error saving last synced reward cycle: %w", err)
		}

		// Add a small delay to avoid overwhelming the API
		time.Sleep(100 * time.Millisecond)
	}

	uc.logger.Infof("Rewards syncing completed up to cycle %d", currentCycle)
	return nil
}

// saveRewardsBatch saves a batch of rewards to the database.
func (uc *SyncRewards) saveRewardsBatch(ctx context.Context, rewards []model.Reward, cycle int) error {
	if len(rewards) == 0 {
		return nil
	}

	// Process rewards in batches
	for i := 0; i < len(rewards); i += uc.batchSize {
		end := i + uc.batchSize
		if end > len(rewards) {
			end = len(rewards)
		}

		batch := rewards[i:end]
		if err := uc.dbAdapter.SaveRewards(ctx, batch); err != nil {
			return fmt.Errorf("error saving rewards batch (cycle %d, batch %d-%d): %w",
				cycle, i, end-1, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}

	uc.logger.Infof("Successfully saved %d rewards for cycle %d", len(rewards), cycle)
	return nil
}

// withMonitorer wraps the Sync function with monitoring capabilities.
func (uc *SyncRewards) withMonitorer(syncRewards model.SyncFunc, metricsClient metrics.Adapter) model.SyncFunc {
	return func(ctx context.Context) (err error) {
		startTime := time.Now()

		defer func() {
			if metricsClient != nil {
				duration := time.Since(startTime)
				metricsClient.RecordServiceOperation("SyncRewards", "UseCase", duration, err)
			}
		}()

		return syncRewards(ctx)
	}
}
