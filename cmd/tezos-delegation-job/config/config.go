package config

import (
	"github.com/spf13/viper"

	datbasefactory "github.com/tezos-delegation-service/internal/adapter/database/factory"
	metricsfactory "github.com/tezos-delegation-service/internal/adapter/metrics/factory"
	tzktapifactory "github.com/tezos-delegation-service/internal/adapter/tzktapi/factory"
	"github.com/tezos-delegation-service/pkg/logger"
)

// Config represents the application configuration.
type Config struct {
	Server ServerConfig

	DatabaseAdapter datbasefactory.Config `mapstructure:"database"`
	TZKTApiAdapter  tzktapifactory.Config `mapstructure:"tzktapi"`
	Pagination      PaginationConfig      `mapstructure:"pagination"`
	Metrics         metricsfactory.Config `mapstructure:"metrics"`
	Logging         logger.Config         `mapstructure:"logging"`
}

// ServerConfig represents the server configuration.
type ServerConfig struct {
	Port uint16 `mapstructure:"port"`
}

// PaginationConfig represents the pagination configuration.
type PaginationConfig struct {
	Limit uint16 `mapstructure:"limit"`
}

// Load loads the configuration from the specified file.
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
