package memory

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestMetrics_RecordAPIRequest(t *testing.T) {
	type fields struct {
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
			name: "Nominal case",
			fields: fields{
				APIRequestsCount: 0,
			},
			args: args{
				method:       "GET",
				path:         "/api/test",
				status:       "200",
				duration:     time.Second,
				responseSize: 1024,
			},
		},
		{
			name: "Error case - invalid method",
			fields: fields{
				APIRequestsCount: 0,
			},
			args: args{
				method:       "",
				path:         "/api/test",
				status:       "400",
				duration:     time.Second,
				responseSize: 1024,
			},
		},
		{
			name: "Error case - invalid path",
			fields: fields{
				APIRequestsCount: 0,
			},
			args: args{
				method:       "GET",
				path:         "",
				status:       "400",
				duration:     time.Second,
				responseSize: 1024,
			},
		},
		{
			name: "Error case - invalid status",
			fields: fields{
				APIRequestsCount: 0,
			},
			args: args{
				method:       "GET",
				path:         "/api/test",
				status:       "",
				duration:     time.Second,
				responseSize: 1024,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				APIRequestsCount:          tt.fields.APIRequestsCount,
				RepositoryOperationsCount: tt.fields.RepositoryOperationsCount,
				RepositoryErrorsCount:     tt.fields.RepositoryErrorsCount,
				ServiceOperationsCount:    tt.fields.ServiceOperationsCount,
				ServiceErrorsCount:        tt.fields.ServiceErrorsCount,
				TZKTAPIRequestsCount:      tt.fields.TZKTAPIRequestsCount,
				DelegationsSyncCount:      tt.fields.DelegationsSyncCount,
				DelegationsTotal:          tt.fields.DelegationsTotal,
				DelegationsAmount:         tt.fields.DelegationsAmount,
				DelegationsFetched:        tt.fields.DelegationsFetched,
			}
			m.RecordAPIRequest(tt.args.method, tt.args.path, tt.args.status, tt.args.duration, tt.args.responseSize)
		})
	}
}

func TestMetrics_RecordDelegationsFetched(t *testing.T) {
	type fields struct {
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
	type args struct {
		count int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Nominal case",
			fields: fields{
				DelegationsFetched: 0,
			},
			args: args{
				count: 5,
			},
		},
		{
			name: "Error case - negative count",
			fields: fields{
				DelegationsFetched: 0,
			},
			args: args{
				count: -1,
			},
		},
		{
			name: "Error case - zero count",
			fields: fields{
				DelegationsFetched: 0,
			},
			args: args{
				count: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				APIRequestsCount:          tt.fields.APIRequestsCount,
				RepositoryOperationsCount: tt.fields.RepositoryOperationsCount,
				RepositoryErrorsCount:     tt.fields.RepositoryErrorsCount,
				ServiceOperationsCount:    tt.fields.ServiceOperationsCount,
				ServiceErrorsCount:        tt.fields.ServiceErrorsCount,
				TZKTAPIRequestsCount:      tt.fields.TZKTAPIRequestsCount,
				DelegationsSyncCount:      tt.fields.DelegationsSyncCount,
				DelegationsTotal:          tt.fields.DelegationsTotal,
				DelegationsAmount:         tt.fields.DelegationsAmount,
				DelegationsFetched:        tt.fields.DelegationsFetched,
			}
			m.RecordDelegationsFetched(tt.args.count)
		})
	}
}

func TestMetrics_RecordDelegationsSync(t *testing.T) {
	type fields struct {
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
	type args struct {
		syncType string
		count    int
		amount   float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Nominal case",
			fields: fields{
				DelegationsSyncCount: 0,
				DelegationsTotal:     0,
				DelegationsAmount:    0.0,
			},
			args: args{
				syncType: "full",
				count:    10,
				amount:   100.0,
			},
		},
		{
			name: "Error case - negative count",
			fields: fields{
				DelegationsSyncCount: 0,
				DelegationsTotal:     0,
				DelegationsAmount:    0.0,
			},
			args: args{
				syncType: "full",
				count:    -5,
				amount:   50.0,
			},
		},
		{
			name: "Error case - zero amount",
			fields: fields{
				DelegationsSyncCount: 0,
				DelegationsTotal:     0,
				DelegationsAmount:    0.0,
			},
			args: args{
				syncType: "full",
				count:    5,
				amount:   0.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				APIRequestsCount:          tt.fields.APIRequestsCount,
				RepositoryOperationsCount: tt.fields.RepositoryOperationsCount,
				RepositoryErrorsCount:     tt.fields.RepositoryErrorsCount,
				ServiceOperationsCount:    tt.fields.ServiceOperationsCount,
				ServiceErrorsCount:        tt.fields.ServiceErrorsCount,
				TZKTAPIRequestsCount:      tt.fields.TZKTAPIRequestsCount,
				DelegationsSyncCount:      tt.fields.DelegationsSyncCount,
				DelegationsTotal:          tt.fields.DelegationsTotal,
				DelegationsAmount:         tt.fields.DelegationsAmount,
				DelegationsFetched:        tt.fields.DelegationsFetched,
			}
			m.RecordDelegationsSync(tt.args.syncType, tt.args.count, tt.args.amount)
		})
	}
}

func TestMetrics_RecordRepositoryOperation(t *testing.T) {
	type fields struct {
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
	type args struct {
		operation string
		repoType  string
		duration  time.Duration
		err       error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Nominal case",
			fields: fields{
				RepositoryOperationsCount: 0,
				RepositoryErrorsCount:     0,
			},
			args: args{
				operation: "insert",
				repoType:  "main",
				duration:  time.Second,
				err:       nil,
			},
		},
		{
			name: "Error case - operation failed",
			fields: fields{
				RepositoryOperationsCount: 0,
				RepositoryErrorsCount:     0,
			},
			args: args{
				operation: "insert",
				repoType:  "main",
				duration:  time.Second,
				err:       errors.New("insert error"),
			},
		},
		{
			name: "Error case - invalid repoType",
			fields: fields{
				RepositoryOperationsCount: 0,
				RepositoryErrorsCount:     0,
			},
			args: args{
				operation: "insert",
				repoType:  "",
				duration:  time.Second,
				err:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				APIRequestsCount:          tt.fields.APIRequestsCount,
				RepositoryOperationsCount: tt.fields.RepositoryOperationsCount,
				RepositoryErrorsCount:     tt.fields.RepositoryErrorsCount,
				ServiceOperationsCount:    tt.fields.ServiceOperationsCount,
				ServiceErrorsCount:        tt.fields.ServiceErrorsCount,
				TZKTAPIRequestsCount:      tt.fields.TZKTAPIRequestsCount,
				DelegationsSyncCount:      tt.fields.DelegationsSyncCount,
				DelegationsTotal:          tt.fields.DelegationsTotal,
				DelegationsAmount:         tt.fields.DelegationsAmount,
				DelegationsFetched:        tt.fields.DelegationsFetched,
			}
			m.RecordRepositoryOperation(tt.args.operation, tt.args.repoType, tt.args.duration, tt.args.err)
		})
	}
}

func TestMetrics_RecordServiceOperation(t *testing.T) {
	type fields struct {
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
	type args struct {
		operation   string
		serviceType string
		duration    time.Duration
		err         error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Nominal case",
			fields: fields{
				ServiceOperationsCount: 0,
				ServiceErrorsCount:     0,
			},
			args: args{
				operation:   "process",
				serviceType: "payment",
				duration:    time.Second,
				err:         nil,
			},
		},
		{
			name: "Error case - operation failed",
			fields: fields{
				ServiceOperationsCount: 0,
				ServiceErrorsCount:     0,
			},
			args: args{
				operation:   "process",
				serviceType: "payment",
				duration:    time.Second,
				err:         errors.New("process error"),
			},
		},
		{
			name: "Error case - invalid serviceType",
			fields: fields{
				ServiceOperationsCount: 0,
				ServiceErrorsCount:     0,
			},
			args: args{
				operation:   "process",
				serviceType: "",
				duration:    time.Second,
				err:         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				APIRequestsCount:          tt.fields.APIRequestsCount,
				RepositoryOperationsCount: tt.fields.RepositoryOperationsCount,
				RepositoryErrorsCount:     tt.fields.RepositoryErrorsCount,
				ServiceOperationsCount:    tt.fields.ServiceOperationsCount,
				ServiceErrorsCount:        tt.fields.ServiceErrorsCount,
				TZKTAPIRequestsCount:      tt.fields.TZKTAPIRequestsCount,
				DelegationsSyncCount:      tt.fields.DelegationsSyncCount,
				DelegationsTotal:          tt.fields.DelegationsTotal,
				DelegationsAmount:         tt.fields.DelegationsAmount,
				DelegationsFetched:        tt.fields.DelegationsFetched,
			}
			m.RecordServiceOperation(tt.args.operation, tt.args.serviceType, tt.args.duration, tt.args.err)
		})
	}
}

func TestMetrics_RecordTZKTAPIRequest(t *testing.T) {
	type fields struct {
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
	type args struct {
		endpoint string
		duration time.Duration
		success  bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Nominal case",
			fields: fields{
				TZKTAPIRequestsCount: 0,
			},
			args: args{
				endpoint: "/api/tzkt",
				duration: time.Second,
				success:  true,
			},
		},
		{
			name: "Error case - invalid endpoint",
			fields: fields{
				TZKTAPIRequestsCount: 0,
			},
			args: args{
				endpoint: "",
				duration: time.Second,
				success:  false,
			},
		},
		{
			name: "Error case - request failed",
			fields: fields{
				TZKTAPIRequestsCount: 0,
			},
			args: args{
				endpoint: "/api/tzkt",
				duration: time.Second,
				success:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				APIRequestsCount:          tt.fields.APIRequestsCount,
				RepositoryOperationsCount: tt.fields.RepositoryOperationsCount,
				RepositoryErrorsCount:     tt.fields.RepositoryErrorsCount,
				ServiceOperationsCount:    tt.fields.ServiceOperationsCount,
				ServiceErrorsCount:        tt.fields.ServiceErrorsCount,
				TZKTAPIRequestsCount:      tt.fields.TZKTAPIRequestsCount,
				DelegationsSyncCount:      tt.fields.DelegationsSyncCount,
				DelegationsTotal:          tt.fields.DelegationsTotal,
				DelegationsAmount:         tt.fields.DelegationsAmount,
				DelegationsFetched:        tt.fields.DelegationsFetched,
			}
			m.RecordTZKTAPIRequest(tt.args.endpoint, tt.args.duration, tt.args.success)
		})
	}
}

func TestNew(t *testing.T) {
	want := &Metrics{
		APIRequestsCount:          0,
		RepositoryOperationsCount: 0,
		RepositoryErrorsCount:     0,
		ServiceOperationsCount:    0,
		ServiceErrorsCount:        0,
		TZKTAPIRequestsCount:      0,
		DelegationsSyncCount:      0,
		DelegationsTotal:          0,
		DelegationsAmount:         0.0,
		DelegationsFetched:        0,
	}
	if got := New(); !reflect.DeepEqual(got, want) {
		t.Errorf("New() = %v, want %v", got, want)
	}
}
