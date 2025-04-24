package proxy

import (
	"context"
	"time"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/model"
)

// TelemetryWrapper is a wrapper for a database adapter that records telemetry metrics.
type TelemetryWrapper struct {
	metrics  metrics.Adapter
	db       database.Adapter
	implType string
}

// New creates a new TelemetryWrapper for a given database adapter.
func New(db database.Adapter, implType string, metrics metrics.Adapter) database.Adapter {
	if db == nil {
		return nil
	}
	return &TelemetryWrapper{
		metrics:  metrics,
		db:       db,
		implType: implType,
	}
}

// Ping checks the database connection and records metrics.
func (w *TelemetryWrapper) Ping() error {
	startTime := time.Now()
	err := w.db.Ping()
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("Ping", w.implType, duration, err)
	}

	return err
}

// GetLatestDelegation retrieves the latest delegation and records metrics.
func (w *TelemetryWrapper) GetLatestDelegation(ctx context.Context) (*model.Delegation, error) {
	startTime := time.Now()
	delegation, err := w.db.GetLatestDelegation(ctx)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("GetLatestDelegation", w.implType, duration, err)
	}

	return delegation, err
}

// GetDelegations retrieves delegations with pagination and records metrics.
func (w *TelemetryWrapper) GetDelegations(ctx context.Context, page uint32, limit, year uint16, maxDelegationID uint64) ([]model.Delegation, error) {
	startTime := time.Now()
	delegations, err := w.db.GetDelegations(ctx, page, limit, year, maxDelegationID)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("GetDelegations", w.implType, duration, err)
	}

	return delegations, err
}

// GetOperations retrieves operations with pagination and records metrics.
func (w *TelemetryWrapper) GetOperations(ctx context.Context, fromDate, toDate int64, page, limit uint16, operationType model.OperationType, wallet, baker model.WalletAddress) ([]model.Operation, error) {
	startTime := time.Now()
	delegations, err := w.db.GetOperations(ctx, fromDate, toDate, page, limit, operationType, wallet, baker)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("GetOperations", w.implType, duration, err)
	}

	return delegations, err
}

// GetRewards retrieves rewards for a given wallet and baker within a date range and records metrics.
func (w *TelemetryWrapper) GetRewards(ctx context.Context, fromDate, toDate int64, wallet, baker model.WalletAddress) ([]model.Reward, error) {
	startTime := time.Now()
	rewards, err := w.db.GetRewards(ctx, fromDate, toDate, wallet, baker)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("GetRewards", w.implType, duration, err)
	}

	return rewards, err
}

// GetHighestBlockLevel retrieves the highest block level and records metrics.
func (w *TelemetryWrapper) GetHighestBlockLevel(ctx context.Context) (uint64, error) {
	startTime := time.Now()
	level, err := w.db.GetHighestBlockLevel(ctx)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("GetHighestBlockLevel", w.implType, duration, err)
	}

	return level, err
}

// SaveDelegation saves a single delegation and records metrics.
func (w *TelemetryWrapper) SaveDelegation(ctx context.Context, delegation *model.Delegation) error {
	startTime := time.Now()
	err := w.db.SaveDelegation(ctx, delegation)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("SaveDelegation", w.implType, duration, err)
	}

	return err
}

// SaveAccount saves a single account and records metrics.
func (w *TelemetryWrapper) SaveAccount(ctx context.Context, account model.Account) error {
	startTime := time.Now()
	err := w.db.SaveAccount(ctx, account)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("SaveAccount", w.implType, duration, err)
	}

	return err
}

// SaveDelegations saves multiple delegations and records metrics.
func (w *TelemetryWrapper) SaveDelegations(ctx context.Context, delegations []*model.Delegation) error {
	startTime := time.Now()
	err := w.db.SaveDelegations(ctx, delegations)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("SaveDelegations", w.implType, duration, err)

		// Record business metrics
		if err == nil {
			amount := 0.0
			for _, d := range delegations {
				amount += d.Amount
			}
			w.metrics.RecordDelegationsSync("repository", len(delegations), amount)
		}
	}

	return err
}

// Close closes the repository and records metrics.
func (w *TelemetryWrapper) Close() error {
	startTime := time.Now()
	err := w.db.Close()
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("Close", w.implType, duration, err)
	}

	return err
}
