package usecase

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/model"
)

// syncRewards handles business logic for syncing delegations.
type syncRewards struct {
	batchSize      int
	dbAdapter      database.Adapter
	logger         *logrus.Entry
	tzktApiAdapter tzktapi.Adapter
}

// NewSyncRewardsFunc creates a new instance of syncRewards.
func NewSyncRewardsFunc(tzktAdapter tzktapi.Adapter, dbAdapter database.Adapter, metricsClient metrics.Adapter, logger *logrus.Entry) model.SyncFunc {
	uc := &syncRewards{
		batchSize:      1000,
		dbAdapter:      dbAdapter,
		logger:         logger.WithField("usecase", "sync_rewards"),
		tzktApiAdapter: tzktAdapter,
	}
	return uc.withMonitorer(uc.SyncRewards, metricsClient)
}

// SyncRewards syncs delegations from the TzKT API to the database.
func (uc *syncRewards) SyncRewards(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
	}

	return nil
}

// withMonitorer wraps the SyncRewards function with monitoring capabilities.
func (uc *syncRewards) withMonitorer(syncRewards model.SyncFunc, metricsClient metrics.Adapter) model.SyncFunc {
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
