package noop

import (
	"time"
)

// Metrics implements the metrics.MetricsClient interface with no-op operations
type Metrics struct{}

// New creates a new no-op metrics client.
func New() *Metrics {
	return &Metrics{}
}

// RecordAPIRequest is a no-op implementation.
func (m *Metrics) RecordAPIRequest(method, path, status string, duration time.Duration, responseSize int) {
}

// RecordRepositoryOperation is a no-op implementation.
func (m *Metrics) RecordRepositoryOperation(operation, repoType string, duration time.Duration, err error) {
}

// RecordServiceOperation is a no-op implementation.
func (m *Metrics) RecordServiceOperation(operation, serviceType string, duration time.Duration, err error) {
}

// RecordTZKTAPIRequest is a no-op implementation.
func (m *Metrics) RecordTZKTAPIRequest(endpoint string, duration time.Duration, success bool) {}

// RecordDelegationsSync is a no-op implementation.
func (m *Metrics) RecordDelegationsSync(syncType string, count int, amount float64) {}

// RecordDelegationsFetched is a no-op implementation.
func (m *Metrics) RecordDelegationsFetched(count int) {}
