package noop

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestMetrics_RecordAPIRequest(t *testing.T) {
	type args struct {
		method       string
		path         string
		status       string
		duration     time.Duration
		responseSize int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nominal case",
			args: args{
				method:       "GET",
				path:         "/api/test",
				status:       "200",
				duration:     100 * time.Millisecond,
				responseSize: 512,
			},
		},
		{
			name: "error case - invalid method",
			args: args{
				method:       "",
				path:         "/api/test",
				status:       "200",
				duration:     100 * time.Millisecond,
				responseSize: 512,
			},
		},
		{
			name: "error case - invalid status",
			args: args{
				method:       "GET",
				path:         "/api/test",
				status:       "",
				duration:     100 * time.Millisecond,
				responseSize: 512,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{}
			m.RecordAPIRequest(tt.args.method, tt.args.path, tt.args.status, tt.args.duration, tt.args.responseSize)
		})
	}
}

func TestMetrics_RecordDelegationsFetched(t *testing.T) {
	type args struct {
		count int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nominal case",
			args: args{
				count: 10,
			},
		},
		{
			name: "error case - negative count",
			args: args{
				count: -1,
			},
		},
		{
			name: "error case - zero count",
			args: args{
				count: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{}
			m.RecordDelegationsFetched(tt.args.count)
		})
	}
}

func TestMetrics_RecordDelegationsSync(t *testing.T) {
	type args struct {
		syncType string
		count    int
		amount   float64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nominal case",
			args: args{
				syncType: "full",
				count:    5,
				amount:   100.0,
			},
		},
		{
			name: "error case - negative count",
			args: args{
				syncType: "full",
				count:    -1,
				amount:   100.0,
			},
		},
		{
			name: "error case - zero amount",
			args: args{
				syncType: "full",
				count:    5,
				amount:   0.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{}
			m.RecordDelegationsSync(tt.args.syncType, tt.args.count, tt.args.amount)
		})
	}
}

func TestMetrics_RecordRepositoryOperation(t *testing.T) {
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
				operation: "clone",
				repoType:  "git",
				duration:  200 * time.Millisecond,
				err:       nil,
			},
		},
		{
			name: "error case - empty operation",
			args: args{
				operation: "",
				repoType:  "git",
				duration:  200 * time.Millisecond,
				err:       nil,
			},
		},
		{
			name: "error case - empty repoType",
			args: args{
				operation: "clone",
				repoType:  "",
				duration:  200 * time.Millisecond,
				err:       nil,
			},
		},
		{
			name: "error case - non-nil error",
			args: args{
				operation: "clone",
				repoType:  "git",
				duration:  200 * time.Millisecond,
				err:       errors.New("some error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{}
			m.RecordRepositoryOperation(tt.args.operation, tt.args.repoType, tt.args.duration, tt.args.err)
		})
	}
}

func TestMetrics_RecordServiceOperation(t *testing.T) {
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
				duration:    150 * time.Millisecond,
				err:         nil,
			},
		},
		{
			name: "error case - empty operation",
			args: args{
				operation:   "",
				serviceType: "payment",
				duration:    150 * time.Millisecond,
				err:         nil,
			},
		},
		{
			name: "error case - empty serviceType",
			args: args{
				operation:   "process",
				serviceType: "",
				duration:    150 * time.Millisecond,
				err:         nil,
			},
		},
		{
			name: "error case - non-nil error",
			args: args{
				operation:   "process",
				serviceType: "payment",
				duration:    150 * time.Millisecond,
				err:         errors.New("some error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{}
			m.RecordServiceOperation(tt.args.operation, tt.args.serviceType, tt.args.duration, tt.args.err)
		})
	}
}

func TestMetrics_RecordTZKTAPIRequest(t *testing.T) {
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
				endpoint: "/api/tzkt",
				duration: 200 * time.Millisecond,
				success:  true,
			},
		},
		{
			name: "error case - empty endpoint",
			args: args{
				endpoint: "",
				duration: 200 * time.Millisecond,
				success:  true,
			},
		},
		{
			name: "error case - zero duration",
			args: args{
				endpoint: "/api/tzkt",
				duration: 0,
				success:  true,
			},
		},
		{
			name: "error case - unsuccessful request",
			args: args{
				endpoint: "/api/tzkt",
				duration: 200 * time.Millisecond,
				success:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{}
			m.RecordTZKTAPIRequest(tt.args.endpoint, tt.args.duration, tt.args.success)
		})
	}
}

func TestNew(t *testing.T) {
	want := &Metrics{}
	if got := New(); !reflect.DeepEqual(got, want) {
		t.Errorf("New() = %v, want %v", got, want)
	}
}
