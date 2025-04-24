package poller

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/model"
	"github.com/tezos-delegation-service/internal/usecase"
)

// usecases holds the use case functions.
type usecases struct {
	ucSyncDelegations model.SyncFunc
	ucSyncOperations  model.SyncFunc
	ucSyncRewards     model.SyncFunc
}

// Poller is a structure that manages the polling process for Tezos delegations.
type Poller struct {
	dbAdapter database.Adapter
	logger    *logrus.Entry

	maxConsecutiveErrors int

	pollingInterval time.Duration
	pollingCtx      context.Context
	pollingCancel   context.CancelFunc
	pollingWg       *sync.WaitGroup

	tzktAdapter  tzktapi.Adapter
	allSyncFuncs map[string]model.SyncFunc
}

// New creates a new Poller instance with the provided TzKT API adapter, database adapter, polling interval, and logger.
func New(tzktAdapter tzktapi.Adapter, dbAdapter database.Adapter, pollingInterval time.Duration, metricClient metrics.Adapter, logger *logrus.Entry) *Poller {
	uc := usecases{
		ucSyncDelegations: usecase.NewSyncDelegationsFunc(tzktAdapter, dbAdapter, metricClient, logger),
		ucSyncOperations:  usecase.NewSyncOperationsFunc(tzktAdapter, dbAdapter, metricClient, logger),
		ucSyncRewards:     usecase.NewSyncRewardsFunc(tzktAdapter, dbAdapter, metricClient, logger),
	}

	return &Poller{
		dbAdapter:            dbAdapter,
		logger:               logger.WithField("component", "poller"),
		pollingInterval:      pollingInterval,
		pollingWg:            &sync.WaitGroup{},
		maxConsecutiveErrors: 5,
		tzktAdapter:          tzktAdapter,
		allSyncFuncs: map[string]model.SyncFunc{
			"delegations": uc.ucSyncDelegations,
			"operations":  uc.ucSyncOperations,
			"rewards":     uc.ucSyncRewards,
		},
	}
}

// Run starts the polling process for Tezos delegations.
func (p *Poller) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	failedOps := make(map[string]model.SyncFunc)
	if err := p.performMultiSync(ctx, "historical", p.allSyncFuncs, failedOps); err != nil {
		p.logger.WithError(err).Error("Historical sync failed, aborting polling")
		return
	}

	p.logger.Info("Starting delegation poller...")
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()

	var syncMutex sync.Mutex
	consecutiveErrors := 0

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Polling stopped")
			return
		case <-ticker.C:
			if syncMutex.TryLock() {
				go func() {
					defer syncMutex.Unlock()

					failedOps = make(map[string]model.SyncFunc)
					if err := p.performMultiSync(ctx, "regular", p.allSyncFuncs, failedOps); err != nil {
						consecutiveErrors++
						p.logger.WithField("consecutiveErrors", consecutiveErrors).
							WithField("failedOps", len(failedOps)).
							WithError(err).
							Error("Regular sync failed")

						if consecutiveErrors >= p.maxConsecutiveErrors {
							p.logger.Warn("Too many consecutive sync errors, attempting recovery sync for failed operations")
							time.Sleep(time.Minute)

							if len(failedOps) > 0 {
								if recoveryErr := p.performMultiSync(ctx, "recovery", failedOps, nil); recoveryErr == nil {
									consecutiveErrors = 0
									p.logger.Info("Recovery sync successful")
								} else {
									p.logger.WithError(recoveryErr).Error("Recovery sync also failed")
								}
							}
						}
					} else {
						consecutiveErrors = 0
					}
				}()
			} else {
				p.logger.Debug("Skipping sync as previous sync is still running")
			}
		}
	}
}

// performMultiSync performs multiple sync operations concurrently and logs the results.
func (p *Poller) performMultiSync(ctx context.Context, syncType string, syncFuncs map[string]model.SyncFunc, failedOps map[string]model.SyncFunc) error {
	start := time.Now()
	var errs []error

	for name, syncFunc := range syncFuncs {
		if err := syncFunc(ctx); err != nil {
			p.logger.WithError(err).Errorf("Error in %s sync (%s)", syncType, name)
			errs = append(errs, err)

			if failedOps != nil {
				failedOps[name] = syncFunc
			}
		} else {
			p.logger.Infof("%s sync (%s) succeeded", syncType, name)
		}
	}

	p.logger.WithField("duration", time.Since(start)).Infof("Multi-sync %s completed", syncType)

	if len(errs) > 0 {
		return errors.New("one or more sync operations failed")
	}
	return nil
}
