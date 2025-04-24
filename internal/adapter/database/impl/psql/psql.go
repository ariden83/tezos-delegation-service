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

// Config represents database configuration.
type Config struct {
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	User             string `mapstructure:"user"`
	Password         Secret `mapstructure:"password"`
	DBName           string `mapstructure:"dbname"`
	SSLMode          string `mapstructure:"sslmode"`
	TableDelegations string `mapstructure:"table_delegations"`
	TableOperations  string `mapstructure:"table_operations"`
	TableRewards     string `mapstructure:"table_rewards"`
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
	db               *sqlx.DB
	tableDelegations string
	tableOperations  string
	tableRewards     string
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
		db:               db,
		tableDelegations: cfg.TableDelegations,
		tableOperations:  cfg.TableOperations,
		tableRewards:     cfg.TableRewards,
	}, nil
}

// Ping checks the database connection.
func (p *psql) Ping() error {
	return p.db.Ping()
}

// GetHighestBlockLevel returns the highest block level in the database.
func (p *psql) GetHighestBlockLevel(ctx context.Context) (uint64, error) {
	var level uint64
	err := p.db.GetContext(ctx, &level, "SELECT COALESCE(MAX(level), 0) FROM "+p.tableDelegations)
	return level, err
}

// GetOperations returns delegations with pagination and optional year, operationType, and maxDelegationID filters.
func (p *psql) GetOperations(ctx context.Context, fromDate, toDate int64, page, limit uint16, operationType model.OperationType, wallet, baker model.WalletAddress) ([]model.Operation, error) {
	var operations []model.Operation
	if page < 1 {
		page = 1
	}

	if limit == 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	offset := (page - 1) * limit

	var (
		query       string
		args        []interface{}
		whereClause string
		argIndex    = 1
	)

	if operationType != "" {
		whereClause = "WHERE o.entrypoint = $" + strconv.Itoa(argIndex)
		args = append(args, operationType.String())
		argIndex++
	}

	if wallet != "" {
		if whereClause == "" {
			whereClause = "WHERE "
		} else {
			whereClause += " AND "
		}
		whereClause += "sender.address = $" + strconv.Itoa(argIndex)
		args = append(args, wallet.String())
		argIndex++
	}

	if baker != "" {
		if whereClause == "" {
			whereClause = "WHERE "
		} else {
			whereClause += " AND "
		}
		whereClause += "contract.address = $" + strconv.Itoa(argIndex)
		args = append(args, baker.String())
		argIndex++
	}

	query = `
		SELECT o.id, sender.address as sender_id, contract.address as contract_id, 
			o.entrypoint, o.amount, o.block, o.timestamp, o.status, o.created_at
		FROM ` + p.tableOperations + ` o
		JOIN app.accounts sender ON o.sender_id = sender.id
		JOIN app.accounts contract ON o.contract_id = contract.id
		` + whereClause + `
		ORDER BY o.timestamp DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1) + `
	`
	args = append(args, limit, offset)

	err := p.db.SelectContext(ctx, &operations, query, args...)
	if err != nil {
		return nil, err
	}

	return operations, nil
}

// GetRewards returns rewards for a given wallet and baker within a date range.
func (p *psql) GetRewards(ctx context.Context, fromDate, toDate uint64, wallet, baker model.WalletAddress) ([]model.Reward, error) {
	return nil, nil
}

// GetLatestDelegation returns the latest delegation from the database.
func (p *psql) GetLatestDelegation(ctx context.Context) (*model.Delegation, error) {
	var delegation model.Delegation
	query := `
		SELECT id, delegator, delegate, timestamp, amount, level, created_at
		FROM ` + p.tableDelegations + `
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
func (p *psql) GetDelegations(ctx context.Context, page uint32, limit, year uint16, maxDelegationID uint64) ([]model.Delegation, error) {
	var delegations []model.Delegation

	if page < 1 {
		page = 1
	}

	if limit == 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
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
		SELECT id, delegator, delegate, timestamp, amount, level, created_at
		FROM ` + p.tableDelegations + `
		` + whereClause + `
		ORDER BY timestamp DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1) + `
	`
	args = append(args, limit, offset)

	err := p.db.SelectContext(ctx, &delegations, query, args...)
	if err != nil {
		return nil, err
	}

	return delegations, nil
}

// SaveDelegation saves a delegation to the database.
func (p *psql) SaveDelegation(ctx context.Context, delegation *model.Delegation) error {
	query := `
		INSERT INTO ` + p.tableDelegations + ` (delegator, delegate, timestamp, amount, level)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`
	_, err := p.db.ExecContext(ctx, query, delegation.Delegator, delegation.Delegate, delegation.Timestamp, delegation.Amount, delegation.Level)
	return err
}

// SaveDelegations saves multiple delegations to the database.
func (p *psql) SaveDelegations(ctx context.Context, delegations []*model.Delegation) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO ` + p.tableDelegations + ` (delegator, delegate, timestamp, amount, level)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`

	for _, delegation := range delegations {
		_, err := tx.ExecContext(ctx, query, delegation.Delegator, delegation.Delegate, delegation.Timestamp, delegation.Amount, delegation.Level)
		if err != nil {
			if errRollBack := tx.Rollback(); errRollBack != nil {
				return errors.New("query execution error: " + err.Error() + ", rollback error: " + errRollBack.Error())
			}
			return err
		}
	}

	return tx.Commit()
}

// Close closes the database connection.
func (p *psql) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
