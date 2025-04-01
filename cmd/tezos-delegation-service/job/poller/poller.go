package poller

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/usecase"
)

// Poller is a structure that manages the polling process for Tezos delegations.
type Poller struct {
	dbAdapter database.Adapter
	logger    *logrus.Entry

	pollingInterval time.Duration
	pollingCtx      context.Context
	pollingCancel   context.CancelFunc
	pollingWg       *sync.WaitGroup

	tzktAdapter       tzktapi.Adapter
	ucSyncDelegations usecase.SyncDelegationsFunc
}

// New creates a new Poller instance with the provided TzKT API adapter, database adapter, polling interval, and logger.
func New(tzktAdapter tzktapi.Adapter, dbAdapter database.Adapter, pollingInterval time.Duration, metricClient metrics.Adapter, logger *logrus.Entry) *Poller {
	return &Poller{
		dbAdapter:         dbAdapter,
		logger:            logger.WithField("component", "poller"),
		pollingInterval:   pollingInterval,
		tzktAdapter:       tzktAdapter,
		ucSyncDelegations: usecase.NewSyncDelegationsFunc(tzktAdapter, dbAdapter, metricClient, logger),
	}
}

// Run starts the polling process for Tezos delegations.
func (p *Poller) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	highestLevel, err := p.dbAdapter.GetHighestBlockLevel(ctx)
	if err != nil {
		p.logger.Errorf("Error checking highest block level: %v", err)
		return
	}

	if highestLevel == 0 {
		p.logger.Info("No delegations found in database. Starting historical sync...")
		if err := p.ucSyncDelegations(ctx); err != nil {
			p.logger.Errorf("Error syncing historical delegations: %v", err)
			return
		}
		p.logger.Info("Historical sync completed")
	}

	p.logger.Info("Starting delegation poller...")
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Polling stopped")
			return
		case <-ticker.C:
			p.StartPolling(ctx)
		}
	}
}

// StartPolling starts polling for new delegations.
func (p *Poller) StartPolling(ctx context.Context) {
	p.StopPolling()

	p.pollingCtx, p.pollingCancel = context.WithCancel(ctx)

	p.pollingWg.Add(1)
	go func() {
		defer p.pollingWg.Done()

		ticker := time.NewTicker(p.pollingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := p.ucSyncDelegations(p.pollingCtx); err != nil {
					p.logger.Errorf("Error syncing delegations: %v", err)
				}
			case <-p.pollingCtx.Done():
				return
			}
		}
	}()
}

// StopPolling stops polling for new delegations.
func (p *Poller) StopPolling() {
	if p.pollingCancel != nil {
		p.pollingCancel()
		p.pollingWg.Wait()
		p.pollingCancel = nil
	}
}
