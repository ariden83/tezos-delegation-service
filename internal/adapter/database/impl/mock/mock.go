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
func (m *Mock) GetOperations(ctx context.Context, page, limit uint16, operationType model.OperationType, wallet, baker model.WalletAddress) ([]model.Operation, error) {
	args := m.Called(ctx, page, limit, operationType, wallet, baker)
	return args.Get(0).([]model.Operation), args.Error(1)
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

// Close closes the database connection.
func (m *Mock) Close() error {
	args := m.Called()
	return args.Error(0)
}
