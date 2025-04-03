package psql

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver (keeping for backward compatibility)
)

// var execCommand = exec.Command

// initConnection initializes the database connection.
func initConnection(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, string(cfg.Password), cfg.DBName, cfg.SSLMode)

	if os.Getenv("GO_TESTING") == "1" {
		return &sqlx.DB{}, nil
	}

	/*if err := runSqitchMigrations(cfg); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}*/

	db, err := sqlx.Connect(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

/*
// runSqitchMigrations runs the database migrations using sqitch.
func runSqitchMigrations(cfg Config) error {
	// Check if the migration file exists
	if _, err := os.Stat(cfg.DBMigrateFile); os.IsNotExist(err) {
		fmt.Printf("Warning: Migration file %s does not exist, skipping migrations\n", cfg.DBMigrateFile)
		return nil // Skip migrations if file doesn't exist
	}

	cmd := execCommand(cfg.DBMigrateFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables directly from the config
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DB_HOST=%s", cfg.Host),
		fmt.Sprintf("DB_PORT=%d", cfg.Port),
		fmt.Sprintf("DB_NAME=%s", cfg.DBName),
		fmt.Sprintf("DB_USER=%s", cfg.User),
		fmt.Sprintf("DB_PASSWORD=%s", cfg.Password),
	)

	return cmd.Run()
} */
