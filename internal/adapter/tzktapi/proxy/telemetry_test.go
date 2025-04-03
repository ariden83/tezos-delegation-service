package proxy

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/adapter/metrics"
	metricsmemory "github.com/tezos-delegation-service/internal/adapter/metrics/impl/memory"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	tzktapimock "github.com/tezos-delegation-service/internal/adapter/tzktapi/impl/mock"
	"github.com/tezos-delegation-service/internal/model"
)

var stubTZKTDelegationResponse = model.TzktDelegationResponse{
	model.TzktDelegation{
		Type:      "delegation",
		ID:        123456,
		Level:     1000,
		Timestamp: time.Now(),
		Block:     "BLockHash123",
		Hash:      "TxHash123",
		Counter:   1,
		Sender: model.TzktAddress{
			Address: "tz1SenderAddress",
			Alias:   "SenderAlias",
		},
		GasLimit: 10000,
		GasUsed:  9000,
		BakerFee: 100,
		Amount:   1000000,
		Delegate: model.TzktDelegate{
			Address: "tz1DelegateAddress",
			Alias:   "DelegateAlias",
		},
		PrevDelegate: &model.TzktDelegate{
			Address: "tz1PrevDelegateAddress",
			Alias:   "PrevDelegateAlias",
		},
		Status: "applied",
		Errors: []model.TzktError{
			{Type: "temporary"},
		},
		Originated: []model.TzktOriginated{
			{
				Address:  "tz1OriginatedAddress",
				TypeHash: 123,
				CodeHash: 456,
				Tzips:    []string{"FA1.2"},
			},
		},
	},
}

func Test_New(t *testing.T) {
	providedTZKTAPI := tzktapimock.New()
	providedMetricsClient := metricsnoop.New()

	type args struct {
		adapter       tzktapi.Adapter
		implType      string
		metricsClient metrics.Adapter
	}
	tests := []struct {
		name string
		args args
		want tzktapi.Adapter
	}{
		{
			name: "nominal case",
			args: args{
				adapter:       providedTZKTAPI,
				implType:      "testImpl",
				metricsClient: providedMetricsClient,
			},
			want: &TelemetryWrapper{
				adapter:  providedTZKTAPI,
				implType: "testImpl",
				metrics:  providedMetricsClient,
			},
		},
		{
			name: "nil adapter",
			args: args{
				adapter:       nil,
				implType:      "testImpl",
				metricsClient: providedMetricsClient,
			},
			want: &TelemetryWrapper{
				adapter:  nil,
				implType: "testImpl",
				metrics:  providedMetricsClient,
			},
		},
		{
			name: "nil metrics client",
			args: args{
				adapter:       providedTZKTAPI,
				implType:      "testImpl",
				metricsClient: nil,
			},
			want: &TelemetryWrapper{
				adapter:  providedTZKTAPI,
				implType: "testImpl",
				metrics:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.adapter, tt.args.implType, tt.args.metricsClient); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TelemetryWrapper_FetchDelegations(t *testing.T) {
	type fields struct {
		adapter  tzktapi.Adapter
		implType string
		metrics  metrics.Adapter
	}
	type args struct {
		ctx    context.Context
		limit  int
		offset int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TzktDelegationResponse
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				adapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegations", mock.Anything, 10, 0).
						Return(stubTZKTDelegationResponse, nil)
					return m
				}(),
				implType: "testImpl",
				metrics:  metricsmemory.New(),
			},
			args: args{
				ctx:    context.Background(),
				limit:  10,
				offset: 0,
			},
			want:    stubTZKTDelegationResponse,
			wantErr: false,
		},
		{
			name: "error case - adapter returns error",
			fields: fields{
				adapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegations", mock.Anything, 10, 0).
						Return(model.TzktDelegationResponse{}, errors.New("adapter error"))
					return m
				}(),
				implType: "testImpl",
				metrics:  metricsnoop.New(),
			},
			args: args{
				ctx:    context.Background(),
				limit:  10,
				offset: 0,
			},
			want:    model.TzktDelegationResponse{},
			wantErr: true,
		},
		{
			name: "error case - context canceled",
			fields: fields{
				adapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegations", mock.Anything, 10, 0).
						Return(model.TzktDelegationResponse{}, context.Canceled)
					return m
				}(),
				implType: "testImpl",
				metrics:  metricsnoop.New(),
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				limit:  10,
				offset: 0,
			},
			want:    model.TzktDelegationResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				adapter:  tt.fields.adapter,
				implType: tt.fields.implType,
				metrics:  tt.fields.metrics,
			}
			got, err := w.FetchDelegations(tt.args.ctx, tt.args.limit, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchDelegations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchDelegations() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TelemetryWrapper_FetchDelegationsFromLevel(t *testing.T) {
	type fields struct {
		adapter  tzktapi.Adapter
		implType string
		metrics  metrics.Adapter
	}
	type args struct {
		ctx   context.Context
		level uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TzktDelegationResponse
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				adapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegationsFromLevel", mock.Anything, uint64(100)).
						Return(stubTZKTDelegationResponse, nil)
					return m
				}(),
				implType: "testImpl",
				metrics:  metricsmemory.New(),
			},
			args: args{
				ctx:   context.Background(),
				level: 100,
			},
			want:    stubTZKTDelegationResponse,
			wantErr: false,
		},
		{
			name: "error case - adapter returns error",
			fields: fields{
				adapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegationsFromLevel", mock.Anything, uint64(100)).
						Return(model.TzktDelegationResponse{}, errors.New("adapter error"))
					return m
				}(),
				implType: "testImpl",
				metrics:  metricsnoop.New(),
			},
			args: args{
				ctx:   context.Background(),
				level: 100,
			},
			want:    model.TzktDelegationResponse{},
			wantErr: true,
		},
		{
			name: "error case - context canceled",
			fields: fields{
				adapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegationsFromLevel", mock.Anything, uint64(100)).
						Return(model.TzktDelegationResponse{}, context.Canceled)
					return m
				}(),
				implType: "testImpl",
				metrics:  metricsmemory.New(),
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				level: 100,
			},
			want:    model.TzktDelegationResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				adapter:  tt.fields.adapter,
				implType: tt.fields.implType,
				metrics:  tt.fields.metrics,
			}
			got, err := w.FetchDelegationsFromLevel(tt.args.ctx, tt.args.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchDelegationsFromLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchDelegationsFromLevel() got = %v, want %v", got, tt.want)
			}
		})
	}
}
