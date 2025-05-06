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

	// GetLastSyncedRewardCycle returns the last synced reward cycle.
	GetLastSyncedRewardCycle(ctx context.Context) (int, error)

	// GetActiveDelegators returns a list of active delegators.
	GetActiveDelegators(ctx context.Context) ([]model.WalletAddress, error)

	// GetBakerForDelegatorAtCycle returns the baker for a delegator at a specific cycle.
	GetBakerForDelegatorAtCycle(ctx context.Context, delegator model.WalletAddress, cycle int) (model.WalletAddress, error)

	// SaveAccount saves an account to the repository.
	SaveAccount(ctx context.Context, account model.Account) error

	// SaveAccounts saves multiple accounts to the repository.
	SaveAccounts(ctx context.Context, account []model.Account) error

	// SaveDelegation saves a delegation to the repository.
	SaveDelegation(ctx context.Context, delegation *model.Delegation) error

	// SaveStakingPools saves multiple staking pools to the repository.
	SaveStakingPools(ctx context.Context, stakingPools []model.StakingPool) error

	// SaveDelegations saves multiple delegations to the repository.
	SaveDelegations(ctx context.Context, delegations []*model.Delegation) error

	// SaveRewards saves multiple rewards to the repository.
	SaveRewards(ctx context.Context, rewards []model.Reward) error

	// SaveLastSyncedRewardCycle saves the last synced reward cycle.
	SaveLastSyncedRewardCycle(ctx context.Context, cycle int) error

	// Close closes the database connection.
	Close() error
}
