package psql

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver (keeping for backward compatibility)
)

var execCommand = exec.Command

// initConnection initializes the database connection.
func initConnection(cfg Config) (*sqlx.DB, error) {

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	if err := runSqitchMigrations(cfg); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	db, err := sqlx.Connect(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// runSqitchMigrations runs the database migrations using sqitch.
func runSqitchMigrations(cfg Config) error {
	cmd := execCommand("../../../../../scripts/db_migrate.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Extract host
	if idx := strings.Index(dsn, "@"); idx > 0 {
		host := dsn[idx+1:]
		if idx := strings.Index(host, ":"); idx > 0 {
			host = host[:idx]
		} else if idx := strings.Index(host, "/"); idx > 0 {
			host = host[:idx]
		}
		cmd.Env = append(os.Environ(), fmt.Sprintf("DB_HOST=%s", host))
	}

	// Extract port
	if idx := strings.Index(dsn, ":"); idx > 0 {
		port := dsn[idx+1:]
		if idx := strings.Index(port, "/"); idx > 0 {
			port = port[:idx]
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("DB_PORT=%s", port))
	}

	// Extract database name
	if idx := strings.LastIndex(dsn, "/"); idx > 0 {
		dbName := dsn[idx+1:]
		if idx := strings.Index(dbName, "?"); idx > 0 {
			dbName = dbName[:idx]
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("DB_NAME=%s", dbName))
	}

	// Extract username and password
	if idx := strings.Index(dsn, "://"); idx > 0 {
		credentials := dsn[idx+3:]
		if idx := strings.Index(credentials, "@"); idx > 0 {
			credentials = credentials[:idx]
			parts := strings.Split(credentials, ":")
			if len(parts) > 0 {
				cmd.Env = append(cmd.Env, fmt.Sprintf("DB_USER=%s", parts[0]))
			}
			if len(parts) > 1 {
				cmd.Env = append(cmd.Env, fmt.Sprintf("DB_PASSWORD=%s", parts[1]))
			}
		}
	}

	return cmd.Run()
}
