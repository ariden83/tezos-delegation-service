package factory

import (
	"fmt"

	"github.com/tezos-delegation-service/internal/adapter/database"
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
	Impl Implementation `mapstructure:"impl"`
	PSQL *psql.Config   `mapstructure:"psql"`
}

// New creates a new repository factory.
func New(cfg Config, metricsClient metrics.Adapter) (database.Adapter, error) {
	var (
		adapter database.Adapter
		err     error
	)

	switch cfg.Impl {
	case ImplPSQL:
		if cfg.PSQL == nil {
			return nil, fmt.Errorf("PSQL config is required for SQL implementation")
		}
		adapter, err = psql.New(*cfg.PSQL)
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
