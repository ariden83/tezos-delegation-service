package prometheus

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func Test_Metrics_RecordAPIRequest(t *testing.T) {
	type fields struct {
		APIRequestsTotal   *prometheus.CounterVec
		APIRequestDuration *prometheus.HistogramVec
		APIResponseSize    *prometheus.HistogramVec
	}
	type args struct {
		method       string
		path         string
		status       string
		duration     time.Duration
		responseSize int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "nominal case",
			args: args{
				method:       "GET",
				path:         "/test",
				status:       "200",
				duration:     time.Second,
				responseSize: 1024,
			},
		},
		{
			name: "error case - invalid status",
			args: args{
				method:       "POST",
				path:         "/test",
				status:       "500",
				duration:     2 * time.Second,
				responseSize: 2048,
			},
		},
		{
			name: "error case - invalid method",
			args: args{
				method:       "INVALID",
				path:         "/test",
				status:       "400",
				duration:     500 * time.Millisecond,
				responseSize: 512,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				APIRequestsTotal:   prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_api_requests_total"}, []string{"method", "path", "status"}),
				APIRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "test_api_request_duration_seconds"}, []string{"method", "path"}),
				APIResponseSize:    prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "test_api_response_size_bytes"}, []string{"method", "path"}),
			}
			m.RecordAPIRequest(tt.args.method, tt.args.path, tt.args.status, tt.args.duration, tt.args.responseSize)
		})
	}
}

func Test_Metrics_RecordDelegationsFetched(t *testing.T) {
	reg := prometheus.NewRegistry()
	promauto.With(reg).NewCounter(prometheus.CounterOpts{Name: "test_delegations_fetched_total"})

	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		{
			name:    "nominal case - positive value",
			count:   10,
			wantErr: false,
		},
		{
			name:    "borderline case - zero value",
			count:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			counter := promauto.With(reg).NewCounter(prometheus.CounterOpts{Name: "test_delegations_fetched_total"})

			m := &Metrics{
				DelegationsFetched: counter,
			}

			func() {
				defer func() {
					if r := recover(); r != nil && !tt.wantErr {
						t.Errorf("RecordDelegationsFetched() a paniqué: %v", r)
					}
				}()
				m.RecordDelegationsFetched(tt.count)
			}()
		})
	}
}

func Test_Metrics_RecordDelegationsSync(t *testing.T) {
	type args struct {
		syncType string
		count    int
		amount   float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nominal case",
			args: args{
				syncType: "full",
				count:    10,
				amount:   1000.0,
			},
			wantErr: false,
		},
		{
			name: "error case - negative count",
			args: args{
				syncType: "full",
				count:    -5,
				amount:   500.0,
			},
			wantErr: true,
		},
		{
			name: "error case - zero amount",
			args: args{
				syncType: "full",
				count:    5,
				amount:   0.0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				TzktAPIDelegationsSync: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_tzkt_delegations_sync_total"}, []string{"sync_type"}),
				DelegationsTotal:       prometheus.NewCounter(prometheus.CounterOpts{Name: "test_delegations_total"}),
				DelegationsAmount:      prometheus.NewCounter(prometheus.CounterOpts{Name: "test_delegations_amount_total"}),
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						if !tt.wantErr {
							t.Errorf("RecordDelegationsSync() a paniqué: %v", r)
						}
					} else if tt.wantErr {
						t.Errorf("RecordDelegationsSync() n'a pas paniqué alors que c'était attendu")
					}
				}()
				m.RecordDelegationsSync(tt.args.syncType, tt.args.count, tt.args.amount)
			}()
		})
	}
}

func Test_Metrics_RecordRepositoryOperation(t *testing.T) {
	type args struct {
		operation string
		repoType  string
		duration  time.Duration
		err       error
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nominal case",
			args: args{
				operation: "read",
				repoType:  "sql",
				duration:  time.Second,
				err:       nil,
			},
		},
		{
			name: "error case - operation failed",
			args: args{
				operation: "write",
				repoType:  "nosql",
				duration:  2 * time.Second,
				err:       errors.New("write error"),
			},
		},
		{
			name: "error case - invalid repository type",
			args: args{
				operation: "delete",
				repoType:  "invalid",
				duration:  500 * time.Millisecond,
				err:       errors.New("invalid repository type"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				RepositoryOperationsTotal:   prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_repository_operations_total"}, []string{"operation", "repository_type"}),
				RepositoryOperationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "test_repository_operation_duration_seconds"}, []string{"operation", "repository_type"}),
				RepositoryErrors:            prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_repository_errors_total"}, []string{"operation", "repository_type", "error_type"}),
			}
			m.RecordRepositoryOperation(tt.args.operation, tt.args.repoType, tt.args.duration, tt.args.err)
		})
	}
}

func Test_Metrics_RecordServiceOperation(t *testing.T) {
	type args struct {
		operation   string
		serviceType string
		duration    time.Duration
		err         error
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nominal case",
			args: args{
				operation:   "process",
				serviceType: "payment",
				duration:    time.Second,
				err:         nil,
			},
		},
		{
			name: "error case - operation failed",
			args: args{
				operation:   "process",
				serviceType: "payment",
				duration:    2 * time.Second,
				err:         errors.New("process error"),
			},
		},
		{
			name: "error case - invalid service type",
			args: args{
				operation:   "process",
				serviceType: "invalid",
				duration:    500 * time.Millisecond,
				err:         errors.New("invalid service type"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				ServiceOperationsTotal:   prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_service_operations_total"}, []string{"operation", "service_type"}),
				ServiceOperationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "test_service_operation_duration_seconds"}, []string{"operation", "service_type"}),
				ServiceErrors:            prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_service_errors_total"}, []string{"operation", "service_type", "error_type"}),
			}
			m.RecordServiceOperation(tt.args.operation, tt.args.serviceType, tt.args.duration, tt.args.err)
		})
	}
}

func Test_Metrics_RecordTZKTAPIRequest(t *testing.T) {
	type args struct {
		endpoint string
		duration time.Duration
		success  bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nominal case",
			args: args{
				endpoint: "delegations",
				duration: time.Second,
				success:  true,
			},
		},
		{
			name: "error case - request failed",
			args: args{
				endpoint: "delegations",
				duration: 2 * time.Second,
				success:  false,
			},
		},
		{
			name: "error case - invalid endpoint",
			args: args{
				endpoint: "invalid",
				duration: 500 * time.Millisecond,
				success:  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				TzktAPIRequestsTotal:   prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_tzkt_api_requests_total"}, []string{"endpoint", "status"}),
				TzktAPIRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "test_tzkt_api_request_duration_seconds"}, []string{"endpoint"}),
			}
			m.RecordTZKTAPIRequest(tt.args.endpoint, tt.args.duration, tt.args.success)
		})
	}
}

func Test_New(t *testing.T) {
	defaultRegisterer := prometheus.DefaultRegisterer
	defaultRegistry := prometheus.DefaultGatherer
	defer func() {
		prometheus.DefaultRegisterer = defaultRegisterer
		prometheus.DefaultGatherer = defaultRegistry
	}()

	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	got := New()

	t.Run("verification of names and labels", func(t *testing.T) {
		assertCounterVecConfig(t, got.APIRequestsTotal, "tezos_delegation_api_requests_total", []string{"method", "path", "status"})

		assertCounterVecConfig(t, got.RepositoryOperationsTotal, "tezos_delegation_repository_operations_total", []string{"operation", "repository_type"})
		assertCounterVecConfig(t, got.ServiceOperationsTotal, "tezos_delegation_service_operations_total", []string{"operation", "service_type"})
		assertCounterVecConfig(t, got.TzktAPIRequestsTotal, "tezos_delegation_tzkt_api_requests_total", []string{"endpoint", "status"})

		assertCounterConfig(t, got.DelegationsTotal, "tezos_delegation_delegations_total")
		assertCounterConfig(t, got.DelegationsAmount, "tezos_delegation_amount_total")
		assertCounterConfig(t, got.DelegationsFetched, "tezos_delegation_fetched_total")
	})
}

// Helpers to check metrics configurations.
func assertCounterVecConfig(t *testing.T, vec *prometheus.CounterVec, expectedName string, expectedLabels []string) {
	t.Helper()
	desc := vec.WithLabelValues(make([]string, len(expectedLabels))...).Desc()

	if !strings.Contains(desc.String(), expectedName) {
		t.Errorf("Counter name attendu %v non trouvé dans %v", expectedName, desc.String())
	}

	for _, label := range expectedLabels {
		if !strings.Contains(desc.String(), label) {
			t.Errorf("Label %v non trouvé dans %v", label, desc.String())
		}
	}
}

// assertCounterConfig checks the configuration of a counter.
func assertCounterConfig(t *testing.T, counter prometheus.Counter, expectedName string) {
	t.Helper()
	desc := counter.Desc()
	if !strings.Contains(desc.String(), expectedName) {
		t.Errorf("Counter name attendu %v non trouvé dans %v", expectedName, desc.String())
	}
}
