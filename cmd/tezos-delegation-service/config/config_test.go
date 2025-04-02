package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tezos-delegation-service/internal/adapter/database/factory"
)

func Test_Load(t *testing.T) {
	content := `
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
	tmpFile, err := os.CreateTemp("", "config.yaml")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpFile.Name()))
	}()

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	cfg, err := Load(tmpFile.Name())
	assert.NoError(t, err)

	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, false, cfg.Logging.EnableFile)
	assert.Equal(t, "/var/log/tezos-delegation-service.log", cfg.Logging.FilePath)
	assert.Equal(t, false, cfg.Logging.Graylog.Enabled)
	assert.Equal(t, "graylog.example.com", cfg.Logging.Graylog.URL)
	assert.Equal(t, 12201, cfg.Logging.Graylog.Port)
	assert.Equal(t, "tezos-delegation-service", cfg.Logging.Graylog.Facility)

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, factory.ImplPSQL, cfg.DatabaseAdapter.Impl)
	assert.Equal(t, "localhost", cfg.DatabaseAdapter.PSQL.Host)
	assert.Equal(t, 5432, cfg.DatabaseAdapter.PSQL.Port)
	assert.Equal(t, "root", cfg.DatabaseAdapter.PSQL.User)
	assert.Equal(t, "password", cfg.DatabaseAdapter.PSQL.Password)
	assert.Equal(t, "delegations", cfg.DatabaseAdapter.PSQL.DBName)
	assert.Equal(t, "disable", cfg.DatabaseAdapter.PSQL.SSLMode)

	assert.Equal(t, "https://api.test.com", cfg.TZKTApiAdapter.API.URL)
	assert.Equal(t, time.Duration(30), cfg.TZKTApiAdapter.PollingInterval)
	assert.Equal(t, 50, cfg.Pagination.Limit)
}

func TestLoad_Error(t *testing.T) {
	cfg, err := Load("non_existent_file.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := `
server:
  port: 9090
invalid
database:
  driver: sqlite3
`
	tmpFile, err := os.CreateTemp("", "invalid_config.yaml")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpFile.Name()))
	}()

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	cfg, err := Load(tmpFile.Name())
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
