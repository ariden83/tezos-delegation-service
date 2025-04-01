package factory

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi/impl/api"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi/impl/mock"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi/proxy"
)

// Implementation defines the type of adapter to use
type Implementation string

const (
	// ImplAPI is the real implementation of the adapter that makes actual API calls.
	ImplAPI Implementation = "api"

	// ImplMock is the mock implementation of the adapter for testing.
	ImplMock Implementation = "mock"
)

// String returns the string representation of the Implementation.
func (i Implementation) String() string {
	return string(i)
}

// Config represents the TzKT API adapter configuration.
type Config struct {
	PollingInterval time.Duration  `mapstructure:"polling_interval"`
	Impl            Implementation `mapstructure:"impl"`
	API             api.Config     `mapstructure:"api"`
}

// New creates a new TzKT API adapter based on the configuration.
func New(cfg Config, metricsClient metrics.Adapter, logger *logrus.Entry) (tzktapi.Adapter, error) {
	var adapter tzktapi.Adapter
	var err error

	switch cfg.Impl {
	case ImplAPI:
		adapter, err = api.New(cfg.API, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create real TzKT API adapter: %w", err)
		}

	case ImplMock:
		adapter = mock.New()

	default:
		return nil, fmt.Errorf("unsupported TzKT API adapter implementation: %s", cfg.Impl)
	}

	return proxy.New(adapter, cfg.Impl.String(), metricsClient), nil
}
