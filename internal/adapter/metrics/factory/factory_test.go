package factory

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/adapter/metrics/impl/memory"
	"github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	internalprometheus "github.com/tezos-delegation-service/internal/adapter/metrics/impl/prometheus"
)

func Test_Implementation_String(t *testing.T) {
	tests := []struct {
		name string
		i    Implementation
		want string
	}{
		{name: "Prometheus", i: ImplPrometheus, want: "prometheus"},
		{name: "Memory", i: ImplMemory, want: "memory"},
		{name: "Noop", i: ImplNoop, want: "noop"},
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
		cfg Config
	}
	tests := []struct {
		name    string
		args    args
		want    metrics.Adapter
		wantErr bool
	}{
		{
			name:    "Prometheus",
			args:    args{cfg: Config{Impl: ImplPrometheus}},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Memory",
			args:    args{cfg: Config{Impl: ImplMemory}},
			want:    memory.New(),
			wantErr: false,
		},
		{
			name:    "Noop",
			args:    args{cfg: Config{Impl: ImplNoop}},
			want:    noop.New(),
			wantErr: false,
		},
		{
			name:    "Unsupported",
			args:    args{cfg: Config{Impl: "unsupported"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prometheus.DefaultRegisterer = prometheus.NewRegistry()
			got, err := New(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.name == "Prometheus" {
				_, ok := got.(*internalprometheus.Metrics)
				if !ok {
					t.Errorf("New() got type %T, want type *internalprometheus.PrometheusAdapter", got)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}
