package database

import (
	"context"

	"github.com/tezos-delegation-service/internal/model"
)

// Adapter defines the interface for delegation repository operations
type Adapter interface {
	// Ping checks the connection to the database.
	Ping() error

	// SaveDelegation saves a delegation to the repository
	SaveDelegation(ctx context.Context, delegation *model.Delegation) error

	// SaveDelegations saves multiple delegations to the repository
	SaveDelegations(ctx context.Context, delegations []*model.Delegation) error

	// GetLatestDelegation returns the latest delegation from the repository
	GetLatestDelegation(ctx context.Context) (*model.Delegation, error)

	// GetDelegations returns delegations with pagination and optional year and maxDelegationID filters
	GetDelegations(ctx context.Context, page uint32, limit, year uint16, maxDelegationID uint64) ([]model.Delegation, int, error)

	// CountDelegations returns the total count of delegations with optional year filter
	CountDelegations(ctx context.Context, year uint16) (int, error)

	// GetHighestBlockLevel returns the highest block level in the repository.
	GetHighestBlockLevel(ctx context.Context) (uint64, error)

	// Close closes the database connection.
	Close() error
}
