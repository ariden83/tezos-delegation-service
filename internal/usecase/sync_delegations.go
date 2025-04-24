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

// syncDelegations handles business logic for syncing delegations.
type syncDelegations struct {
	batchSizeAPI   int
	batchSizeDB    int
	dbAdapter      database.Adapter
	logger         *logrus.Entry
	tzktApiAdapter tzktapi.Adapter
	maxWorkers     int
}

// NewSyncDelegationsFunc creates a new instance of syncDelegations.
func NewSyncDelegationsFunc(tzktAdapter tzktapi.Adapter, dbAdapter database.Adapter, metricsClient metrics.Adapter, logger *logrus.Entry) model.SyncFunc {
	uc := &syncDelegations{
		batchSizeAPI:   1000,
		batchSizeDB:    100,
		dbAdapter:      dbAdapter,
		logger:         logger.WithField("usecase", "sync_delegations"),
		maxWorkers:     2,
		tzktApiAdapter: tzktAdapter,
	}
	return uc.withMonitorer(uc.SyncDelegations, metricsClient)
}

// SyncDelegations syncs delegations from the TzKT API to the database.
func (uc *syncDelegations) SyncDelegations(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
	}

	highestLevel, err := uc.dbAdapter.GetHighestBlockLevel(ctx)
	if err != nil {
		highestLevel = 0
	}

	if highestLevel == 0 {
		// If no delegations exist, do a full historical sync (from beginning of Tezos in 2018)
		return uc.syncHistoricalDelegations(ctx)
	}

	return uc.syncIncrementalDelegations(ctx, highestLevel)
}

// syncHistoricalDelegations syncs all historical delegations from 2018 (Tezos launch).
func (uc *syncDelegations) syncHistoricalDelegations(ctx context.Context) error {
	uc.logger.Info("Starting full historical delegations sync from 2018...")
	offset := 0
	totalProcessed := 0
	var lastSavedLevel int64 = 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		delegations, err := uc.tzktApiAdapter.FetchDelegations(ctx, uc.batchSizeAPI, offset)
		if err != nil {
			return fmt.Errorf("error fetching historical delegations (offset %d): %w", offset, err)
		}

		if len(delegations) == 0 {
			break
		}

		modelDelegations := make([]*model.Delegation, 0, len(delegations))
		modelAccounts := map[string]*model.Account{}
		for _, d := range delegations {
			if d.Status != "applied" {
				continue
			}

			modelDelegation := &model.Delegation{
				Delegator: model.WalletAddress(d.Sender.Address),
				Delegate:  model.WalletAddress(d.Delegate.Address),
				Amount:    float64(d.Amount) / 1000000.0, // Convert mutez to tez
				Timestamp: d.Timestamp.Unix(),
				Level:     d.Level,
			}

			modelDelegations = append(modelDelegations, modelDelegation)

			if _, exists := modelAccounts[d.Sender.Address]; !exists {
				modelAccounts[d.Sender.Address] = &model.Account{
					Address: model.WalletAddress(d.Sender.Address),
					Alias:   d.Sender.Alias,
					Type:    model.AccountTypeUser,
				}
			}

			if _, exists := modelAccounts[d.Delegate.Address]; !exists {
				modelAccounts[d.Delegate.Address] = &model.Account{
					Address: model.WalletAddress(d.Delegate.Address),
					Alias:   d.Delegate.Alias,
					Type:    model.AccountTypeDelegate,
				}
			}

			if d.Level > lastSavedLevel {
				lastSavedLevel = d.Level
			}
		}

		if len(modelAccounts) > 0 {
			if err := uc.saveAccountsBatch(ctx, modelAccounts, offset); err != nil {
				return err
			}

			if err := uc.saveStakingPoolsBatch(ctx, modelAccounts, offset); err != nil {
				return err
			}
		}

		if len(modelDelegations) > 0 {
			if err := uc.saveDelegations(ctx, modelDelegations, offset); err != nil {
				return err
			}
			totalProcessed += len(modelDelegations)
		}

		if offset%10000 == 0 {
			uc.logger.Infof("Synced %d historical delegations up to level %d\n", totalProcessed, lastSavedLevel)
		}

		offset += uc.batchSizeAPI

		time.Sleep(100 * time.Millisecond)
	}

	uc.logger.Infof("Historical sync completed. Total delegations: %d\n", totalProcessed)
	return nil
}

// saveAccountsBatch saves a batch of accounts to the database.
func (uc *syncDelegations) saveAccountsBatch(ctx context.Context, modelAccounts map[string]*model.Account, offset int) error {
	accounts := make([]model.Account, 0, len(modelAccounts))
	for _, account := range modelAccounts {
		accounts = append(accounts, *account)
	}

	for i := 0; i < len(accounts); i += uc.batchSizeDB {
		end := i + uc.batchSizeDB
		if end > len(accounts) {
			end = len(accounts)
		}

		batch := accounts[i:end]
		if err := uc.dbAdapter.SaveAccounts(ctx, batch); err != nil {
			return fmt.Errorf("error saving accounts batch (offset %d, batch %d-%d): %w",
				offset, i, end-1, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}

	uc.logger.Infof("Successfully saved %d accounts (offset %d)", len(accounts), offset)
	return nil
}

// saveStakingPoolsBatch saves a batch of staking pools to the database.
func (uc *syncDelegations) saveStakingPoolsBatch(ctx context.Context, modelAccounts map[string]*model.Account, offset int) error {
	stakingPools := make([]model.StakingPool, 0, len(modelAccounts))

	for _, account := range modelAccounts {
		if account.Type == model.AccountTypeDelegate {
			stakingPools = append(stakingPools, model.StakingPool{
				Address:      account.Address,
				Name:         account.Alias,
				StakingToken: "XTZ",
			})
		}
	}

	for i := 0; i < len(stakingPools); i += uc.batchSizeDB {
		end := i + uc.batchSizeDB
		if end > len(stakingPools) {
			end = len(stakingPools)
		}

		batch := stakingPools[i:end]

		if err := uc.dbAdapter.SaveStakingPools(ctx, batch); err != nil {
			return fmt.Errorf("error saving staking pools batch (offset %d, batch %d-%d): %w", offset, i, end-1, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}

	uc.logger.Infof("Successfully saved %d staking pools (offset %d)", len(stakingPools), offset)
	return nil
}

// saveDelegations saves a batch of delegations to the database.
func (uc *syncDelegations) saveDelegations(ctx context.Context, delegations []*model.Delegation, offset int) error {
	for i := 0; i < len(delegations); i += uc.batchSizeDB {
		end := i + uc.batchSizeDB
		if end > len(delegations) {
			end = len(delegations)
		}

		batch := delegations[i:end]
		if err := uc.dbAdapter.SaveDelegations(ctx, batch); err != nil {
			return fmt.Errorf("error saving delegations batch (offset %d, batch %d-%d): %w",
				offset, i, end-1, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}

	uc.logger.Infof("Successfully saved %d delegations (offset %d)", len(delegations), offset)
	return nil
}

// syncIncrementalDelegations syncs delegations from a specific block level
func (uc *syncDelegations) syncIncrementalDelegations(ctx context.Context, level uint64) error {
	uc.logger.Infof("Syncing incremental delegations from level %d\n", level)

	delegations, err := uc.tzktApiAdapter.FetchDelegationsFromLevel(ctx, level)
	if err != nil {
		return fmt.Errorf("error fetching delegations from level %d: %w", level, err)
	}

	modelDelegations := make([]*model.Delegation, 0, len(delegations))
	for _, d := range delegations {
		if d.Status != "applied" {
			continue
		}

		modelDelegation := &model.Delegation{
			Delegator: model.WalletAddress(d.Sender.Address),
			Delegate:  model.WalletAddress(d.Delegate.Address),
			Amount:    float64(d.Amount) / 1000000.0, // Convert mutez to tez
			Timestamp: d.Timestamp.Unix(),
			Level:     d.Level,
		}

		modelDelegations = append(modelDelegations, modelDelegation)
	}

	if len(modelDelegations) > 0 {
		if err := uc.dbAdapter.SaveDelegations(ctx, modelDelegations); err != nil {
			return fmt.Errorf("error saving delegations: %w", err)
		}
		uc.logger.Infof("Synced %d new delegations\n", len(modelDelegations))
	} else {
		uc.logger.Info("No new delegations to sync")
	}

	return nil
}

// withMonitorer wraps the SyncDelegations function with monitoring capabilities.
func (uc *syncDelegations) withMonitorer(syncDelegations model.SyncFunc, metricsClient metrics.Adapter) model.SyncFunc {
	return func(ctx context.Context) (err error) {
		startTime := time.Now()

		defer func() {
			if metricsClient != nil {
				duration := time.Since(startTime)
				metricsClient.RecordServiceOperation("SyncDelegations", "UseCase", duration, err)
			}
		}()

		return syncDelegations(ctx)
	}
}
