package factory

import (
	"fmt"
	"testing"

	"github.com/tezos-delegation-service/internal/adapter/database"
	// "github.com/tezos-delegation-service/internal/adapter/database/impl/memory"
	"github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
	"github.com/tezos-delegation-service/internal/adapter/database/impl/psql"
	"github.com/tezos-delegation-service/internal/adapter/database/proxy"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
)

// Implementation defines the type of repository to use
type Implementation string

const (
	// ImplPSQL is the PostgreSQL implementation of the repository.
	ImplPSQL Implementation = "psql"

	// ImplMemory is the in-memory implementation of the repository.
	ImplMemory Implementation = "memory"
)

// String returns the string representation of the Implementation.
func (i Implementation) String() string {
	return string(i)
}

// Config represents the database configuration.
type Config struct {
	Driver string
	DSN    string
	Impl   Implementation `mapstructure:"impl"`
	psql   *psql.Config   `mapstructure:"psql"`
}

// New creates a new repository factory.
func New(cfg Config, metricsClient metrics.Adapter) (database.Adapter, error) {
	var (
		adapter database.Adapter
		err     error
	)

	switch cfg.Impl {
	case ImplPSQL:
		if cfg.psql == nil {
			return nil, fmt.Errorf("psql config is required for SQL implementation")
		}
		adapter, err = psql.New(*cfg.psql)
		if err != nil {
			return nil, fmt.Errorf("failed to create SQL repository: %w", err)
		}
		// case ImplMemory:
	// 	adapter = memory.New()
	default:
		return nil, fmt.Errorf("unsupported implementation type: %s", cfg.Impl)
	}

	return proxy.New(adapter, cfg.Impl.String(), metricsClient), nil
}

// NewMock returns a new filestorage mock adapter.
func NewMock(t *testing.T) *mock.Mock {
	return mock.New()
}
