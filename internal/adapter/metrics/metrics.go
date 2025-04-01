package metrics

import "time"

// Adapter defines the interface for metrics collection.
type Adapter interface {
	RecordAPIRequest(method, path, status string, duration time.Duration, responseSize int)
	RecordRepositoryOperation(operation, repoType string, duration time.Duration, err error)
	RecordServiceOperation(operation, serviceType string, duration time.Duration, err error)
	RecordTZKTAPIRequest(endpoint string, duration time.Duration, success bool)
	RecordDelegationsSync(syncType string, count int, amount float64)
	RecordDelegationsFetched(count int)
}
