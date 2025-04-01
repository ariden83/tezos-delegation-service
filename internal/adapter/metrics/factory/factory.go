package factory

import (
	"fmt"

	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/metrics/impl/memory"
	"github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	"github.com/tezos-delegation-service/internal/adapter/metrics/impl/prometheus"
)

// Implementation defines the type of metrics implementation to use.
type Implementation string

const (
	// ImplPrometheus is the Prometheus implementation of metrics.
	ImplPrometheus Implementation = "prometheus"

	// ImplMemory is the in-memory implementation of metrics.
	ImplMemory Implementation = "memory"

	// ImplNoop is the no-op implementation of metrics (does nothing).
	ImplNoop Implementation = "noop"
)

// String returns the string representation of the Implementation.
func (i Implementation) String() string {
	return string(i)
}

// Config represents the metrics configuration.
type Config struct {
	Impl Implementation `mapstructure:"impl"`
}

// New creates a new metrics client based on the provided implementation.
func New(cfg Config) (metrics.Adapter, error) {
	switch cfg.Impl {
	case ImplPrometheus:
		return prometheus.New(), nil
	case ImplMemory:
		return memory.New(), nil
	case ImplNoop:
		return noop.New(), nil
	default:
		return nil, fmt.Errorf("unsupported metrics implementation: %s", cfg.Impl)
	}
}
