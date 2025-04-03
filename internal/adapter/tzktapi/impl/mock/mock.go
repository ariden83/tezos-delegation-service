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

// FetchDelegations fetches delegations from the TzKT API.
func (m *Mock) FetchDelegations(ctx context.Context, limit, offset int) (model.TzktDelegationResponse, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).(model.TzktDelegationResponse), args.Error(1)
}

// FetchDelegationsFromLevel fetches delegations from a specific level.
func (m *Mock) FetchDelegationsFromLevel(ctx context.Context, level uint64) (model.TzktDelegationResponse, error) {
	args := m.Called(ctx, level)
	return args.Get(0).(model.TzktDelegationResponse), args.Error(1)
}
