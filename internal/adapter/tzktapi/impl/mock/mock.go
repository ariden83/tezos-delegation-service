package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
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

// FetchDelegations fetches delegations from the TzKT API.
func (m *Mock) FetchDelegations(ctx context.Context, limit uint16, offset int) (model.TzktDelegationResponse, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).(model.TzktDelegationResponse), args.Error(1)
}

// FetchDelegationsFromLevel fetches delegations from a specific level.
func (m *Mock) FetchDelegationsFromLevel(ctx context.Context, level uint64, limit uint8) (model.TzktDelegationResponse, error) {
	args := m.Called(ctx, level)
	return args.Get(0).(model.TzktDelegationResponse), args.Error(1)
}

// FetchOperationsFromTezos fetches operations from the Tezos node.
func (m *Mock) FetchOperationsFromTezos(blockID string) ([]model.Operation, error) {
	args := m.Called(blockID)
	return args.Get(0).([]model.Operation), args.Error(1)
}

// FetchRewardsForBaker fetches rewards for a specific baker from the Tezos node.
func (m *Mock) FetchRewardsForBaker(blockID, bakerAddress string) (model.Reward, error) {
	args := m.Called(blockID, bakerAddress)
	return args.Get(0).(model.Reward), args.Error(1)
}

// FetchWalletInfo fetches wallet information from the Tezos node.
func (m *Mock) FetchWalletInfo(blockID, walletAddress string) (model.WalletInfo, error) {
	args := m.Called(blockID, walletAddress)
	return args.Get(0).(model.WalletInfo), args.Error(1)
}

// FetchStakingOperations fetches staking operations from the TzKT API.
func (m *Mock) FetchStakingOperations(ctx context.Context, filter tzktapi.OperationFilter) ([]model.StakingOperation, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.StakingOperation), args.Error(1)
}

// GetCurrentCycle gets the current cycle from the TzKT API.
func (m *Mock) GetCurrentCycle(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// FetchRewardsForCycle fetches rewards for a specific delegator and baker in a given cycle.
func (m *Mock) FetchRewardsForCycle(ctx context.Context, delegator model.WalletAddress, baker model.WalletAddress, cycle int) ([]model.Reward, error) {
	args := m.Called(ctx, delegator, baker, cycle)
	return args.Get(0).([]model.Reward), args.Error(1)
}
