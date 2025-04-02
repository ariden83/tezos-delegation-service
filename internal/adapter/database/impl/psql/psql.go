package psql

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/model"
)

// Config represents database configuration.
type Config struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// psql implements DelegationRepository using SQL database.
type psql struct {
	db *sqlx.DB
}

// New creates a new SQL delegation repository.
func New(cfg Config) (database.Adapter, error) {
	db, err := initConnection(cfg)
	if err != nil {
		return nil, err
	} else if db == nil {
		return nil, errors.New("failed to initialize database connection")
	}

	return &psql{
		db: db,
	}, nil
}

// Ping checks the database connection.
func (p *psql) Ping() error {
	return p.db.Ping()
}

// SaveDelegation saves a delegation to the database.
func (p *psql) SaveDelegation(ctx context.Context, delegation *model.Delegation) error {
	query := `
		INSERT INTO delegations (delegator, timestamp, amount, level)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`
	_, err := p.db.ExecContext(ctx, query, delegation.Delegator, delegation.Timestamp, delegation.Amount, delegation.Level)
	return err
}

// SaveDelegations saves multiple delegations to the database.
func (p *psql) SaveDelegations(ctx context.Context, delegations []*model.Delegation) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO delegations (delegator, timestamp, amount, level)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`

	for _, delegation := range delegations {
		_, err := tx.ExecContext(ctx, query, delegation.Delegator, delegation.Timestamp, delegation.Amount, delegation.Level)
		if err != nil {
			if errRollBack := tx.Rollback(); errRollBack != nil {
				return errors.New("query execution error: " + err.Error() + ", rollback error: " + errRollBack.Error())
			}
			return err
		}
	}

	return tx.Commit()
}

// GetLatestDelegation returns the latest delegation from the database.
func (p *psql) GetLatestDelegation(ctx context.Context) (*model.Delegation, error) {
	var delegation model.Delegation
	query := `
		SELECT id, delegator, timestamp, amount, level, created_at
		FROM delegations
		ORDER BY level DESC
		LIMIT 1
	`
	err := p.db.GetContext(ctx, &delegation, query)
	if err != nil {
		return nil, err
	}
	return &delegation, nil
}

// GetDelegations returns delegations with pagination and optional year filter.
func (p *psql) GetDelegations(ctx context.Context, page int, limit int, year int) ([]model.Delegation, int, error) {
	var delegations []model.Delegation
	if limit <= 0 {
		limit = 50
	}
	offset := (page - 1) * limit

	var query string
	var args []interface{}

	if year > 0 {
		startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
		endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

		query = `
			SELECT id, delegator, timestamp, amount, level, created_at
			FROM delegations
			WHERE timestamp >= $1 AND timestamp < $2
			ORDER BY timestamp DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{startDate, endDate, limit, offset}
	} else {
		query = `
			SELECT id, delegator, timestamp, amount, level, created_at
			FROM delegations
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	err := p.db.SelectContext(ctx, &delegations, query, args...)
	if err != nil {
		return nil, 0, err
	}

	totalCount, err := p.CountDelegations(ctx, year)
	if err != nil {
		return nil, 0, err
	}

	return delegations, totalCount, nil
}

// CountDelegations returns the total count of delegations with optional year filter.
func (p *psql) CountDelegations(ctx context.Context, year int) (int, error) {
	var count int
	var query string
	var args []interface{}

	if year > 0 {
		startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
		endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

		query = `
			SELECT COUNT(*) 
			FROM delegations
			WHERE timestamp >= $1 AND timestamp < $2
		`
		args = []interface{}{startDate, endDate}
	} else {
		query = `
			SELECT COUNT(*) 
			FROM delegations
		`
	}

	err := p.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetHighestBlockLevel returns the highest block level in the database.
func (p *psql) GetHighestBlockLevel(ctx context.Context) (int64, error) {
	var level int64
	err := p.db.GetContext(ctx, &level, "SELECT COALESCE(MAX(level), 0) FROM delegations")
	return level, err
}

// Close closes the database connection.
func (p *psql) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
