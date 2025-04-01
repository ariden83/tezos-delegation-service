package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contains all the Prometheus metrics.
type Metrics struct {
	// API Metrics
	APIRequestsTotal   *prometheus.CounterVec
	APIRequestDuration *prometheus.HistogramVec
	APIResponseSize    *prometheus.HistogramVec

	// Repository Metrics
	RepositoryOperationsTotal   *prometheus.CounterVec
	RepositoryOperationDuration *prometheus.HistogramVec
	RepositoryErrors            *prometheus.CounterVec

	// Service Metrics
	ServiceOperationsTotal   *prometheus.CounterVec
	ServiceOperationDuration *prometheus.HistogramVec
	ServiceErrors            *prometheus.CounterVec

	// TzKT API Metrics
	TzktAPIRequestsTotal   *prometheus.CounterVec
	TzktAPIRequestDuration *prometheus.HistogramVec
	TzktAPIResponseSize    *prometheus.HistogramVec
	TzktAPIDelegationsSync *prometheus.CounterVec

	// Business Metrics
	DelegationsTotal   prometheus.Counter
	DelegationsAmount  prometheus.Counter
	DelegationsFetched prometheus.Counter
}

// New creates and registers all application metrics.
func New() *Metrics {
	m := &Metrics{
		// API Metrics
		APIRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_delegation_api_requests_total",
				Help: "Total number of API requests",
			},
			[]string{"method", "path", "status"},
		),
		APIRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tezos_delegation_api_request_duration_seconds",
				Help:    "API request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		APIResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tezos_delegation_api_response_size_bytes",
				Help:    "API response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "path"},
		),

		// Repository Metrics
		RepositoryOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_delegation_repository_operations_total",
				Help: "Total number of repository operations",
			},
			[]string{"operation", "repository_type"},
		),
		RepositoryOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tezos_delegation_repository_operation_duration_seconds",
				Help:    "Repository operation duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "repository_type"},
		),
		RepositoryErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_delegation_repository_errors_total",
				Help: "Total number of repository errors",
			},
			[]string{"operation", "repository_type", "error_type"},
		),

		// Service Metrics
		ServiceOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_delegation_service_operations_total",
				Help: "Total number of service operations",
			},
			[]string{"operation", "service_type"},
		),
		ServiceOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tezos_delegation_service_operation_duration_seconds",
				Help:    "Service operation duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "service_type"},
		),
		ServiceErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_delegation_service_errors_total",
				Help: "Total number of service errors",
			},
			[]string{"operation", "service_type", "error_type"},
		),

		// TzKT API Metrics
		TzktAPIRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_delegation_tzkt_api_requests_total",
				Help: "Total number of TzKT API requests",
			},
			[]string{"endpoint", "status"},
		),
		TzktAPIRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tezos_delegation_tzkt_api_request_duration_seconds",
				Help:    "TzKT API request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		),
		TzktAPIResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tezos_delegation_tzkt_api_response_size_bytes",
				Help:    "TzKT API response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"endpoint"},
		),
		TzktAPIDelegationsSync: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tezos_delegation_tzkt_delegations_sync_total",
				Help: "Total number of delegations synced from TzKT API",
			},
			[]string{"sync_type"},
		),

		// Business Metrics
		DelegationsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tezos_delegation_delegations_total",
				Help: "Total number of delegations processed",
			},
		),
		DelegationsAmount: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tezos_delegation_amount_total",
				Help: "Total amount delegated in mutez",
			},
		),
		DelegationsFetched: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tezos_delegation_fetched_total",
				Help: "Total number of delegations fetched from the API",
			},
		),
	}

	return m
}

// RecordAPIRequest records metrics for an API request.
func (m *Metrics) RecordAPIRequest(method, path, status string, duration time.Duration, responseSize int) {
	m.APIRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.APIRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	m.APIResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

// RecordRepositoryOperation records metrics for a repository operation.
func (m *Metrics) RecordRepositoryOperation(operation, repoType string, duration time.Duration, err error) {
	m.RepositoryOperationsTotal.WithLabelValues(operation, repoType).Inc()
	m.RepositoryOperationDuration.WithLabelValues(operation, repoType).Observe(duration.Seconds())

	if err != nil {
		m.RepositoryErrors.WithLabelValues(operation, repoType, "error").Inc()
	}
}

// RecordServiceOperation records metrics for a service operation.
func (m *Metrics) RecordServiceOperation(operation, serviceType string, duration time.Duration, err error) {
	m.ServiceOperationsTotal.WithLabelValues(operation, serviceType).Inc()
	m.ServiceOperationDuration.WithLabelValues(operation, serviceType).Observe(duration.Seconds())

	if err != nil {
		m.ServiceErrors.WithLabelValues(operation, serviceType, "error").Inc()
	}
}

// RecordTZKTAPIRequest records metrics for a TZKT API request.
func (m *Metrics) RecordTZKTAPIRequest(endpoint string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	m.TzktAPIRequestsTotal.WithLabelValues(endpoint, status).Inc()
	m.TzktAPIRequestDuration.WithLabelValues(endpoint).Observe(duration.Seconds())
}

// RecordDelegationsSync records metrics for delegations synced from TZKT API.
func (m *Metrics) RecordDelegationsSync(syncType string, count int, amount float64) {
	m.TzktAPIDelegationsSync.WithLabelValues(syncType).Add(float64(count))
	m.DelegationsTotal.Add(float64(count))
	m.DelegationsAmount.Add(amount)
}

// RecordDelegationsFetched records metrics for delegations fetched from the database.
func (m *Metrics) RecordDelegationsFetched(count int) {
	m.DelegationsFetched.Add(float64(count))
}
