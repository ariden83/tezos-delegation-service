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

// syncOperations handles business logic for syncing delegations.
type syncOperations struct {
	batchSize      int
	dbAdapter      database.Adapter
	logger         *logrus.Entry
	tzktApiAdapter tzktapi.Adapter
}

// NewSyncOperationsFunc creates a new instance of syncOperations.
func NewSyncOperationsFunc(tzktAdapter tzktapi.Adapter, dbAdapter database.Adapter, metricsClient metrics.Adapter, logger *logrus.Entry) model.SyncFunc {
	uc := &syncOperations{
		batchSize:      1000,
		dbAdapter:      dbAdapter,
		logger:         logger.WithField("usecase", "sync_operations"),
		tzktApiAdapter: tzktAdapter,
	}
	return uc.withMonitorer(uc.SyncOperations, metricsClient)
}

// SyncOperations syncs delegations from the TzKT API to the database.
func (uc *syncOperations) SyncOperations(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
	}

	return nil
}

// withMonitorer wraps the SyncOperations function with monitoring capabilities.
func (uc *syncOperations) withMonitorer(syncOperations model.SyncFunc, metricsClient metrics.Adapter) model.SyncFunc {
	return func(ctx context.Context) (err error) {
		startTime := time.Now()

		defer func() {
			if metricsClient != nil {
				duration := time.Since(startTime)
				metricsClient.RecordServiceOperation("SyncOperations", "UseCase", duration, err)
			}
		}()

		return syncOperations(ctx)
	}
}
