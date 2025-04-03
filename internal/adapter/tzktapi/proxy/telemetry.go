package proxy

import (
	"context"
	"time"

	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/model"
)

// TelemetryWrapper wraps a TzKT API adapter with telemetry
type TelemetryWrapper struct {
	adapter  tzktapi.Adapter
	implType string
	metrics  metrics.Adapter
}

// New creates a new telemetry wrapper for a TzKT API adapter
func New(adapter tzktapi.Adapter, implType string, metricsClient metrics.Adapter) tzktapi.Adapter {
	return &TelemetryWrapper{
		adapter:  adapter,
		implType: implType,
		metrics:  metricsClient,
	}
}

// FetchDelegations fetches delegations with telemetry
func (w *TelemetryWrapper) FetchDelegations(ctx context.Context, limit, offset int) (model.TzktDelegationResponse, error) {
	startTime := time.Now()
	endpoint := "delegations"

	result, err := w.adapter.FetchDelegations(ctx, limit, offset)

	if w.metrics != nil {
		w.metrics.RecordTZKTAPIRequest(endpoint, time.Since(startTime), err == nil)
	}

	return result, err
}

// FetchDelegationsFromLevel fetches delegations from a level with telemetry
func (w *TelemetryWrapper) FetchDelegationsFromLevel(ctx context.Context, level uint64) (model.TzktDelegationResponse, error) {
	startTime := time.Now()
	endpoint := "delegations_from_level"

	result, err := w.adapter.FetchDelegationsFromLevel(ctx, level)

	if w.metrics != nil {
		w.metrics.RecordTZKTAPIRequest(endpoint, time.Since(startTime), err == nil)
	}

	return result, err
}
