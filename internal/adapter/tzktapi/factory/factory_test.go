package factory

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/metrics"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi/impl/api"
)

func Test_Implementation_String(t *testing.T) {
	tests := []struct {
		name string
		i    Implementation
		want string
	}{
		{
			name: "Nominal case",
			i:    ImplAPI,
			want: "api",
		},
		{
			name: "Error case - empty implementation",
			i:    Implementation(""),
			want: "",
		},
		{
			name: "Error case - invalid implementation",
			i:    Implementation("invalid"),
			want: "invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_New(t *testing.T) {
	type args struct {
		cfg           Config
		metricsClient metrics.Adapter
		logger        *logrus.Entry
	}
	tests := []struct {
		name    string
		args    args
		want    tzktapi.Adapter
		wantErr bool
	}{
		{
			name: "Nominal case - API implementation",
			args: args{
				cfg: Config{
					PollingInterval: 10 * time.Second,
					Impl:            ImplAPI,
					API:             api.Config{URL: "https://api.tzkt.io"},
				},
				metricsClient: metricsnoop.New(),
				logger:        logrus.NewEntry(logrus.New()),
			},
			wantErr: false,
		},
		{
			name: "Error case - Unsupported implementation",
			args: args{
				cfg: Config{
					PollingInterval: 10 * time.Second,
					Impl:            Implementation("unsupported"),
					API:             api.Config{URL: "https://api.tzkt.io"},
				},
				metricsClient: metricsnoop.New(),
				logger:        logrus.NewEntry(logrus.New()),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error case - API creation failure",
			args: args{
				cfg: Config{
					PollingInterval: 10 * time.Second,
					Impl:            ImplAPI,
					API:             api.Config{URL: ""},
				},
				metricsClient: metricsnoop.New(),
				logger:        logrus.NewEntry(logrus.New()),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg, tt.args.metricsClient, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("New() got = nil, want non-nil adapter")
			}
			if tt.wantErr && got != tt.want {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}
