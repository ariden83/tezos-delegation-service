package tzktapi

import (
	"context"

	"github.com/tezos-delegation-service/internal/model"
)

// Adapter defines the operations available in the TzKT API adapter.
type Adapter interface {
	// FetchDelegations fetches delegations from the TzKT API.
	FetchDelegations(ctx context.Context, limit uint16, offset int) (model.TzktDelegationResponse, error)

	// FetchDelegationsFromLevel fetches delegations from a specific level.
	FetchDelegationsFromLevel(ctx context.Context, level uint64, limit uint8) (model.TzktDelegationResponse, error)

	// FetchOperationsFromTezos fetches operations from the Tezos node.
	FetchOperationsFromTezos(blockID string) ([]model.Operation, error)

	// FetchRewardsForBaker fetches rewards for a specific baker from the Tezos node.
	FetchRewardsForBaker(blockID, bakerAddress string) (model.Reward, error)

	// FetchWalletInfo fetches wallet information from the Tezos node.
	FetchWalletInfo(blockID, walletAddress string) (model.WalletInfo, error)

	// FetchStakingOperations fetches staking operations from the TzKT API.
	FetchStakingOperations(ctx context.Context, filter OperationFilter) ([]model.StakingOperation, error)

	// GetCurrentCycle gets the current cycle from the TzKT API.
	GetCurrentCycle(ctx context.Context) (int, error)

	// FetchRewardsForCycle fetches rewards for a specific delegator and baker in a given cycle.
	FetchRewardsForCycle(ctx context.Context, delegator model.WalletAddress, baker model.WalletAddress, cycle int) ([]model.Reward, error)
}
