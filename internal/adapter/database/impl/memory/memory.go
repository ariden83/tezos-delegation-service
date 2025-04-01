package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/model"
)

// Memory implements the database.Adapter interface using an in-memory store.
type Memory struct {
	delegations []*model.Delegation
	nextID      int64
	mu          *sync.RWMutex
}

// New creates a new in-memory delegation repository.
func New() database.Adapter {
	return &Memory{
		delegations: make([]*model.Delegation, 0),
		nextID:      1,
	}
}

// Ping checks the connection to the in-memory store (no-op for mock implementation).
func (m *Memory) Ping() error {
	return nil
}

// SaveDelegation saves a delegation to the in-memory store.
func (m *Memory) SaveDelegation(_ context.Context, delegation *model.Delegation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delegation.ID = m.nextID
	delegation.CreatedAt = time.Now()
	m.nextID++
	m.delegations = append(m.delegations, delegation)
	return nil
}

// SaveDelegations saves multiple delegations to the in-memory store.
func (m *Memory) SaveDelegations(_ context.Context, delegations []*model.Delegation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, delegation := range delegations {
		delegation.ID = m.nextID
		delegation.CreatedAt = time.Now()
		m.nextID++
		m.delegations = append(m.delegations, delegation)
	}
	return nil
}

// GetLatestDelegation returns the latest delegation from the in-memory store.
func (m *Memory) GetLatestDelegation(_ context.Context) (*model.Delegation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.delegations) == 0 {
		return nil, nil
	}

	sort.Slice(m.delegations, func(i, j int) bool {
		return m.delegations[i].Level > m.delegations[j].Level
	})

	latest := *m.delegations[0]
	return &latest, nil
}

// GetDelegations returns delegations with pagination and optional year filter.
func (m *Memory) GetDelegations(_ context.Context, page, limit, year int) ([]model.Delegation, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}
	offset := (page - 1) * limit

	var filtered []*model.Delegation
	if year > 0 {
		startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

		for _, d := range m.delegations {
			delegTime := time.Unix(d.Timestamp, 0)
			if delegTime.After(startDate) && delegTime.Before(endDate) {
				filtered = append(filtered, d)
			}
		}
	} else {
		filtered = m.delegations
	}

	totalCount := len(filtered)

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp > filtered[j].Timestamp
	})

	var result []model.Delegation
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	if offset < len(filtered) {
		for _, d := range filtered[offset:end] {
			result = append(result, *d)
		}
	}

	return result, totalCount, nil
}

// CountDelegations returns the total count of delegations with optional year filter.
func (m *Memory) CountDelegations(_ context.Context, year int) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if year <= 0 {
		return len(m.delegations), nil
	}

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	count := 0
	for _, d := range m.delegations {
		delegTime := time.Unix(d.Timestamp, 0)
		if delegTime.After(startDate) && delegTime.Before(endDate) {
			count++
		}
	}

	return count, nil
}

// GetHighestBlockLevel returns the highest block level in the in-memory store.
func (m *Memory) GetHighestBlockLevel(_ context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.delegations) == 0 {
		return 0, nil
	}

	var highestLevel int64
	for _, d := range m.delegations {
		if d.Level > highestLevel {
			highestLevel = d.Level
		}
	}

	return highestLevel, nil
}

// Close closes the in-memory store (no-op for mock implementation).
func (m *Memory) Close() error {
	return nil
}
