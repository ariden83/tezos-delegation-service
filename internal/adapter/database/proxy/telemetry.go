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
func (w *TelemetryWrapper) GetDelegations(ctx context.Context, page int, limit int, year int) ([]model.Delegation, int, error) {
	startTime := time.Now()
	delegations, totalCount, err := w.db.GetDelegations(ctx, page, limit, year)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("GetDelegations", w.implType, duration, err)
	}

	return delegations, totalCount, err
}

// CountDelegations counts the number of delegations for a given year and records metrics.
func (w *TelemetryWrapper) CountDelegations(ctx context.Context, year int) (int, error) {
	startTime := time.Now()
	count, err := w.db.CountDelegations(ctx, year)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("CountDelegations", w.implType, duration, err)
	}

	return count, err
}

// GetHighestBlockLevel retrieves the highest block level and records metrics.
func (w *TelemetryWrapper) GetHighestBlockLevel(ctx context.Context) (int64, error) {
	startTime := time.Now()
	level, err := w.db.GetHighestBlockLevel(ctx)
	duration := time.Since(startTime)

	if w.metrics != nil {
		w.metrics.RecordRepositoryOperation("GetHighestBlockLevel", w.implType, duration, err)
	}

	return level, err
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
