package psql

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/model"
)

const (
	// TableDelegations is the name of the delegations table.
	TableDelegations = "delegations"
)

// Config represents database configuration.
type Config struct {
	// DBMigrateFile string `mapstructure:"db_migrate_file"`
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password Secret `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type text interface {
	~[]byte | ~string
}

func conceal[T text](s T) string {
	if len(s) == 0 {
		return ""
	}
	return "******"
}

// Secret allows to avoid displaying secret string values in logs for instance.
type Secret string

// String implements Stringer.
func (s Secret) String() string {
	return conceal(s)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (s Secret) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
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
		INSERT INTO ` + TableDelegations + ` (delegator, timestamp, amount, level)
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
		INSERT INTO ` + TableDelegations + ` (delegator, timestamp, amount, level)
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
		FROM ` + TableDelegations + `
		ORDER BY level DESC
		LIMIT 1
	`
	err := p.db.GetContext(ctx, &delegation, query)
	if err != nil {
		return nil, err
	}
	return &delegation, nil
}

// GetDelegations returns delegations with pagination and optional year and maxDelegationID filters.
func (p *psql) GetDelegations(ctx context.Context, page uint32, limit, year uint16, maxDelegationID uint64) ([]model.Delegation, int, error) {
	var delegations []model.Delegation
	if limit == 0 {
		limit = 50
	}
	offset := (page - 1) * uint32(limit)

	var (
		query       string
		args        []interface{}
		whereClause string
		argIndex    = 1
	)

	applyMaxIDFilter := maxDelegationID > 0 && page > 1 && (year == 0 || int(year) == time.Now().Year())

	if year > 0 {
		startDate := time.Date(int(year), 1, 1, 0, 0, 0, 0, time.UTC).Unix()
		endDate := time.Date(int(year)+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
		whereClause = "WHERE timestamp >= $" + strconv.Itoa(argIndex) + " AND timestamp < $" + strconv.Itoa(argIndex+1)
		args = append(args, startDate, endDate)
		argIndex += 2

		if applyMaxIDFilter {
			whereClause += " AND id <= $" + strconv.Itoa(argIndex)
			args = append(args, maxDelegationID)
			argIndex++
		}
	} else {
		if applyMaxIDFilter {
			whereClause = "WHERE id <= $" + strconv.Itoa(argIndex)
			args = append(args, maxDelegationID)
			argIndex++
		}
	}

	query = `
		SELECT id, delegator, timestamp, amount, level, created_at
		FROM ` + TableDelegations + `
		` + whereClause + `
		ORDER BY timestamp DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1) + `
	`
	args = append(args, limit, offset)

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
func (p *psql) CountDelegations(ctx context.Context, year uint16) (int, error) {
	var count int
	var query string
	var args []interface{}

	if year > 0 {
		startDate := time.Date(int(year), 1, 1, 0, 0, 0, 0, time.UTC).Unix()
		endDate := time.Date(int(year)+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

		query = `
			SELECT COUNT(*) 
			FROM ` + TableDelegations + `
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
func (p *psql) GetHighestBlockLevel(ctx context.Context) (uint64, error) {
	var level uint64
	err := p.db.GetContext(ctx, &level, "SELECT COALESCE(MAX(level), 0) FROM "+TableDelegations)
	return level, err
}

// Close closes the database connection.
func (p *psql) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
