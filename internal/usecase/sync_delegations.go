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
	batchSizeDB             int
	batchSizeAPIHistoric    uint16
	batchSizeAPIIncremental uint8
	dbAdapter               database.Adapter
	logger                  *logrus.Entry
	tzktApiAdapter          tzktapi.Adapter
	maxWorkers              int
	isHistoricalSyncDone    bool
}

// NewSyncDelegationsFunc creates a new instance of syncDelegations.
func NewSyncDelegationsFunc(tzktAdapter tzktapi.Adapter, dbAdapter database.Adapter, metricsClient metrics.Adapter, logger *logrus.Entry) model.SyncFunc {
	uc := &syncDelegations{
		batchSizeDB:             100,
		batchSizeAPIHistoric:    1000,
		batchSizeAPIIncremental: 150,
		dbAdapter:               dbAdapter,
		logger:                  logger.WithField("usecase", "sync_delegations"),
		maxWorkers:              2,
		tzktApiAdapter:          tzktAdapter,
		isHistoricalSyncDone:    false,
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
		uc.isHistoricalSyncDone = false
		uc.logger.Info("Resetting historical sync flag because database appears empty")
	}

	if highestLevel == 0 || !uc.isHistoricalSyncDone {
		// Si aucune délégation n'existe ou si la synchronisation historique n'est pas terminée
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

		delegations, err := uc.tzktApiAdapter.FetchDelegations(ctx, uc.batchSizeAPIHistoric, offset)
		if err != nil {
			if err.Error() == "EOF" {
				uc.logger.Info("Reached end of delegations data")
				break
			}
			return fmt.Errorf("error fetching historical delegations (offset %d): %w", offset, err)
		}

		if len(delegations) == 0 {
			break
		}

		for _, d := range delegations {
			if d.Level > lastSavedLevel {
				lastSavedLevel = d.Level
			}
		}

		if err := uc.processDelegations(ctx, delegations, offset); err != nil {
			return err
		}

		totalProcessed += len(delegations)

		if offset%10000 == 0 {
			uc.logger.Infof("Synced %d historical delegations up to level %d\n", totalProcessed, lastSavedLevel)
		}

		offset += int(uc.batchSizeAPIHistoric)

		time.Sleep(100 * time.Millisecond)
	}

	uc.logger.Infof("Historical sync completed. Total delegations: %d\n", totalProcessed)
	uc.isHistoricalSyncDone = true
	return nil
}

// processDelegations processes and saves delegations to the database.
func (uc *syncDelegations) processDelegations(ctx context.Context, delegations model.TzktDelegationResponse, offset int) error {
	modelDelegations := make([]*model.Delegation, 0, len(delegations))
	modelAccounts := map[string]*model.Account{}

	for _, d := range delegations {
		if d.Status != "applied" {
			continue
		}

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

		modelDelegation := &model.Delegation{
			Delegator: model.WalletAddress(d.Sender.Address),
			Delegate:  model.WalletAddress(d.Delegate.Address),
			Amount:    float64(d.Amount) / 1000000.0, // Convert mutez to tez
			Timestamp: d.Timestamp.Unix(),
			Level:     d.Level,
		}

		modelDelegations = append(modelDelegations, modelDelegation)
	}

	if len(modelAccounts) > 0 {
		if err := uc.saveAccountsBatch(ctx, modelAccounts, offset); err != nil {
			uc.logger.Warnf("Error saving accounts: %v", err)
		}

		if err := uc.saveStakingPoolsBatch(ctx, modelAccounts, offset); err != nil {
			uc.logger.Warnf("Error saving staking pools: %v", err)
		}
	}

	if len(modelDelegations) > 0 {
		if err := uc.saveDelegations(ctx, modelDelegations, offset); err != nil {
			return fmt.Errorf("error saving delegations: %w", err)
		}
		uc.logger.Infof("Synced %d new delegations (offset: %d)\n", len(modelDelegations), offset)

	} else {
		uc.logger.Info("No new delegations to sync")
	}

	return nil
}

// syncIncrementalDelegations syncs delegations from a specific block level
func (uc *syncDelegations) syncIncrementalDelegations(ctx context.Context, level uint64) error {
	uc.logger.Infof("Syncing incremental delegations from level %d\n", level)

	delegations, err := uc.tzktApiAdapter.FetchDelegationsFromLevel(ctx, level, uc.batchSizeAPIIncremental)
	if err != nil {
		if err.Error() == "EOF" {
			uc.logger.Info("Reached end of delegations data")
			return nil
		}
		return fmt.Errorf("error fetching delegations from level %d: %w", level, err)
	}

	if err := uc.processDelegations(ctx, delegations, 0); err != nil {
		return err
	}

	if len(delegations) >= int(uc.batchSizeAPIIncremental) {
		uc.logger.Warnf("Large number of delegations (%d) detected. Switching to historical sync mode for subsequent data", len(delegations))
		uc.isHistoricalSyncDone = false
	}

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
