package database

import (
	"context"

	"github.com/tezos-delegation-service/internal/model"
)

// Adapter defines the interface for delegation repository operations
type Adapter interface {
	// Ping checks the connection to the database.
	Ping() error

	// GetDelegations returns delegations with pagination and optional year and maxDelegationID filters.
	GetDelegations(ctx context.Context, page uint32, limit, year uint16, maxDelegationID uint64) ([]model.Delegation, error)

	// GetLatestDelegation returns the latest delegation from the repository.
	GetLatestDelegation(ctx context.Context) (*model.Delegation, error)

	// GetHighestBlockLevel returns the highest block level in the repository.
	GetHighestBlockLevel(ctx context.Context) (uint64, error)

	// GetOperations returns operations with pagination and optional filters.
	GetOperations(ctx context.Context, fromDate, toDate int64, page, limit uint16, operationType model.OperationType, wallet, baker model.WalletAddress) ([]model.Operation, error)

	// GetRewards returns rewards for a given wallet and baker within a date range.
	GetRewards(ctx context.Context, fromDate, toDate int64, wallet, baker model.WalletAddress) ([]model.Reward, error)

	// SaveAccount saves an account to the repository.
	SaveAccount(ctx context.Context, account model.Account) error

	// SaveAccounts saves multiple accounts to the repository.
	SaveAccounts(ctx context.Context, account []model.Account) error

	// SaveDelegation saves a delegation to the repository.
	SaveDelegation(ctx context.Context, delegation *model.Delegation) error

	// SaveDelegations saves multiple delegations to the repository.
	SaveDelegations(ctx context.Context, delegations []*model.Delegation) error

	// Close closes the database connection.
	Close() error
}
