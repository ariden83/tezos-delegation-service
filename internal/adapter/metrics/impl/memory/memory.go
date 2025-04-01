package memory

import (
	"time"
)

// Metrics implements the metrics.MetricsClient interface with in-memory counters.
type Metrics struct {
	APIRequestsCount          int
	RepositoryOperationsCount int
	RepositoryErrorsCount     int
	ServiceOperationsCount    int
	ServiceErrorsCount        int
	TZKTAPIRequestsCount      int
	DelegationsSyncCount      int
	DelegationsTotal          int
	DelegationsAmount         float64
	DelegationsFetched        int
}

// New creates a new memory metrics client.
func New() *Metrics {
	return &Metrics{}
}

// RecordAPIRequest records metrics for an API request.
func (m *Metrics) RecordAPIRequest(method, path, status string, duration time.Duration, responseSize int) {
	m.APIRequestsCount++
}

// RecordRepositoryOperation records metrics for a repository operation.
func (m *Metrics) RecordRepositoryOperation(operation, repoType string, duration time.Duration, err error) {
	m.RepositoryOperationsCount++

	if err != nil {
		m.RepositoryErrorsCount++
	}
}

// RecordServiceOperation records metrics for a service operation.
func (m *Metrics) RecordServiceOperation(operation, serviceType string, duration time.Duration, err error) {
	m.ServiceOperationsCount++

	if err != nil {
		m.ServiceErrorsCount++
	}
}

// RecordTZKTAPIRequest records metrics for a TZKT API request.
func (m *Metrics) RecordTZKTAPIRequest(endpoint string, duration time.Duration, success bool) {
	m.TZKTAPIRequestsCount++
}

// RecordDelegationsSync records metrics for delegations synced from TzKT API.
func (m *Metrics) RecordDelegationsSync(syncType string, count int, amount float64) {
	m.DelegationsSyncCount += count
	m.DelegationsTotal += count
	m.DelegationsAmount += amount
}

// RecordDelegationsFetched records metrics for delegations fetched from the database.
func (m *Metrics) RecordDelegationsFetched(count int) {
	m.DelegationsFetched += count
}
