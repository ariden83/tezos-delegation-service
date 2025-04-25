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
}
