package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tezos-delegation-service/internal/adapter/database/factory"
	"github.com/tezos-delegation-service/internal/adapter/database/impl/psql"
)

func Test_Load(t *testing.T) {
	validConfig := `
logging:
  level: info
  format: json
  enable_file: false
  file_path: /var/log/tezos-delegation-service.log
  graylog:
    enabled: false
    url: graylog.example.com
    port: 12201
    facility: tezos-delegation-service

server:
  port: 9090
  
database:
  impl: psql
  psql:
    host: localhost
    port: 5432
    user: "root"
    password: "password"
    dbname: delegations
    sslmode: disable


tzktapi:
  impl: api
  api:
    url: https://api.test.com
  polling_interval: 30
  
pagination:
  limit: 50
`

	invalidConfig := `
server:
  port: 9090
invalid
database:
  driver: sqlite3
`

	testCases := []struct {
		name        string
		content     string
		filePath    string
		createFile  bool
		expectError bool
		validate    func(t *testing.T, cfg *Config)
	}{
		{
			name:        "valid config",
			content:     validConfig,
			createFile:  true,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "info", cfg.Logging.Level)
				assert.Equal(t, "json", cfg.Logging.Format)
				assert.Equal(t, false, cfg.Logging.EnableFile)
				assert.Equal(t, "/var/log/tezos-delegation-service.log", cfg.Logging.FilePath)
				assert.Equal(t, false, cfg.Logging.Graylog.Enabled)
				assert.Equal(t, "graylog.example.com", cfg.Logging.Graylog.URL)
				assert.Equal(t, 12201, cfg.Logging.Graylog.Port)
				assert.Equal(t, "tezos-delegation-service", cfg.Logging.Graylog.Facility)

				assert.Equal(t, 9090, int(cfg.Server.Port))
				assert.Equal(t, factory.ImplPSQL, cfg.DatabaseAdapter.Impl)
				assert.Equal(t, "localhost", cfg.DatabaseAdapter.PSQL.Host)
				assert.Equal(t, 5432, cfg.DatabaseAdapter.PSQL.Port)
				assert.Equal(t, "root", cfg.DatabaseAdapter.PSQL.User)
				assert.Equal(t, psql.Secret("password"), cfg.DatabaseAdapter.PSQL.Password)
				assert.Equal(t, "delegations", cfg.DatabaseAdapter.PSQL.DBName)
				assert.Equal(t, "disable", cfg.DatabaseAdapter.PSQL.SSLMode)

				assert.Equal(t, uint16(50), cfg.Pagination.Limit)
			},
		},
		{
			name:        "non-existent file",
			filePath:    "non_existent_file.yaml",
			createFile:  false,
			expectError: true,
		},
		{
			name:        "invalid YAML",
			content:     invalidConfig,
			createFile:  true,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filePath string

			if tc.createFile {
				prefix := "config"
				if tc.expectError {
					prefix = "invalid_config"
				}
				tmpFile, err := os.CreateTemp("", prefix+".yaml")
				assert.NoError(t, err)
				defer func() {
					assert.NoError(t, os.Remove(tmpFile.Name()))
				}()

				_, err = tmpFile.WriteString(tc.content)
				assert.NoError(t, err)
				err = tmpFile.Close()
				assert.NoError(t, err)

				filePath = tmpFile.Name()
			} else {
				filePath = tc.filePath
			}

			cfg, err := Load(filePath)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				tc.validate(t, cfg)
			}
		})
	}
}
