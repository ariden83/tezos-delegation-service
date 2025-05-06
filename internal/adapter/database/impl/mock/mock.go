package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/model"
)

type Mock struct {
	mock.Mock
}

// New creates a new mock instance for testing.
func New() *Mock {
	var m Mock
	return &m
}

// Ping checks the connection to the database.
func (m *Mock) Ping() error {
	args := m.Called()
	return args.Error(0)
}

// GetLatestDelegation returns the latest delegation from the repository.
func (m *Mock) GetLatestDelegation(ctx context.Context) (*model.Delegation, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Delegation), args.Error(1)
}

// GetDelegations returns delegations with pagination and optional year and maxDelegationID filters.
func (m *Mock) GetDelegations(ctx context.Context, page uint32, limit, year uint16, maxDelegationID uint64) ([]model.Delegation, error) {
	args := m.Called(ctx, page, limit, year, maxDelegationID)
	return args.Get(0).([]model.Delegation), args.Error(1)
}

// GetHighestBlockLevel returns the highest block level in the repository.
func (m *Mock) GetHighestBlockLevel(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetOperations returns operations with pagination and optional filters.
func (m *Mock) GetOperations(ctx context.Context, fromDate, toDate int64, page, limit uint16, operationType model.OperationType, wallet, baker model.WalletAddress) ([]model.Operation, error) {
	args := m.Called(ctx, fromDate, toDate, page, limit, operationType, wallet, baker)
	return args.Get(0).([]model.Operation), args.Error(1)
}

// GetRewards returns rewards for a given wallet and baker within a date range.
func (m *Mock) GetRewards(ctx context.Context, fromDate, toDate int64, wallet, baker model.WalletAddress) ([]model.Reward, error) {
	args := m.Called(ctx, fromDate, toDate, wallet, baker)
	return args.Get(0).([]model.Reward), args.Error(1)
}

// SaveDelegation saves a delegation to the repository.
func (m *Mock) SaveDelegation(ctx context.Context, delegation *model.Delegation) error {
	args := m.Called(ctx, delegation)
	return args.Error(0)
}

// SaveAccount saves a single account to the repository.
func (m *Mock) SaveAccount(ctx context.Context, account model.Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

// SaveAccounts saves multiple accounts to the repository.
func (m *Mock) SaveAccounts(ctx context.Context, accounts []model.Account) error {
	args := m.Called(ctx, accounts)
	return args.Error(0)
}

// SaveDelegations saves multiple delegations to the repository.
func (m *Mock) SaveDelegations(ctx context.Context, delegations []*model.Delegation) error {
	args := m.Called(ctx, delegations)
	return args.Error(0)
}

// SaveStakingPools saves multiple staking pools to the repository.
func (m *Mock) SaveStakingPools(ctx context.Context, stakingPools []model.StakingPool) error {
	args := m.Called(ctx, stakingPools)
	return args.Error(0)
}

// SaveRewards saves multiple rewards to the repository.
func (m *Mock) SaveRewards(ctx context.Context, rewards []model.Reward) error {
	args := m.Called(ctx, rewards)
	return args.Error(0)
}

// GetLastSyncedRewardCycle returns the last synced reward cycle.
func (m *Mock) GetLastSyncedRewardCycle(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// GetActiveDelegators returns a list of active delegators.
func (m *Mock) GetActiveDelegators(ctx context.Context) ([]model.WalletAddress, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.WalletAddress), args.Error(1)
}

// GetBakerForDelegatorAtCycle returns the baker for a delegator at a specific cycle.
func (m *Mock) GetBakerForDelegatorAtCycle(ctx context.Context, delegator model.WalletAddress, cycle int) (model.WalletAddress, error) {
	args := m.Called(ctx, delegator, cycle)
	return args.Get(0).(model.WalletAddress), args.Error(1)
}

// SaveLastSyncedRewardCycle saves the last synced reward cycle.
func (m *Mock) SaveLastSyncedRewardCycle(ctx context.Context, cycle int) error {
	args := m.Called(ctx, cycle)
	return args.Error(0)
}

// Close closes the database connection.
func (m *Mock) Close() error {
	args := m.Called()
	return args.Error(0)
}
