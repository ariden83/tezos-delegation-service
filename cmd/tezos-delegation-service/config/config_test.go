package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Load(t *testing.T) {
	content := `
server:
  port: 9090
  
database:
  driver: sqlite3
  dsn: ./test.db
  repository_type: sql

tzkt:
  api_url: https://api.test.com
  polling_interval: 30
  
pagination:
  limit: 25

service:
  type: real
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

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "sqlite3", cfg.DatabaseAdapter.Driver)
	assert.Equal(t, "./test.db", cfg.DatabaseAdapter.DSN)
	assert.Equal(t, "sql", cfg.DatabaseAdapter.Impl)
	assert.Equal(t, "https://api.test.com", cfg.TZKTApiAdapter.API.URL)
	assert.Equal(t, 30, cfg.TZKTApiAdapter.PollingInterval)
	assert.Equal(t, 25, cfg.Pagination.Limit)
	assert.Equal(t, "real", cfg.Service.Type)
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
  invalid:
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
