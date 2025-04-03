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

// SaveDelegation saves a delegation to the repository.
func (m *Mock) SaveDelegation(ctx context.Context, delegation *model.Delegation) error {
	args := m.Called(ctx, delegation)
	return args.Error(0)
}

// SaveDelegations saves multiple delegations to the repository.
func (m *Mock) SaveDelegations(ctx context.Context, delegations []*model.Delegation) error {
	args := m.Called(ctx, delegations)
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
func (m *Mock) GetDelegations(ctx context.Context, page uint32, limit, year uint16, maxDelegationID uint64) ([]model.Delegation, int, error) {
	args := m.Called(ctx, page, limit, year, maxDelegationID)
	return args.Get(0).([]model.Delegation), args.Int(1), args.Error(2)
}

// CountDelegations returns the total count of delegations with optional year filter.
func (m *Mock) CountDelegations(ctx context.Context, year uint16) (int, error) {
	args := m.Called(ctx, year)
	return args.Int(0), args.Error(1)
}

// GetHighestBlockLevel returns the highest block level in the repository.
func (m *Mock) GetHighestBlockLevel(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// Close closes the database connection.
func (m *Mock) Close() error {
	args := m.Called()
	return args.Error(0)
}
