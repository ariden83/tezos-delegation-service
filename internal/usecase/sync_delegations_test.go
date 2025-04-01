package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/adapter/database"
	databasemock "github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	tzktapimock "github.com/tezos-delegation-service/internal/adapter/tzktapi/impl/mock"
	"github.com/tezos-delegation-service/internal/model"
)

func Test_NewSyncDelegationsFunc(t *testing.T) {
	providedTZKTAPI := tzktapimock.New()
	mockDbAdapter := databasemock.New()
	providedMetricsClient := metricsnoop.New()

	type args struct {
		tzktAdapter   tzktapi.Adapter
		dbAdapter     database.Adapter
		metricsClient metrics.Adapter
		logger        *logrus.Entry
	}
	tests := []struct {
		name string
		args args
		want SyncDelegationsFunc
	}{
		{
			name: "nominal case",
			args: args{
				tzktAdapter:   providedTZKTAPI,
				dbAdapter:     mockDbAdapter,
				metricsClient: providedMetricsClient,
				logger:        logrus.NewEntry(logrus.New()),
			},
			want: func() SyncDelegationsFunc {
				return NewSyncDelegationsFunc(providedTZKTAPI, mockDbAdapter, providedMetricsClient, logrus.NewEntry(logrus.New()))
			}(),
		},
		{
			name: "error case - nil tzktAdapter",
			args: args{
				tzktAdapter:   nil,
				dbAdapter:     mockDbAdapter,
				metricsClient: providedMetricsClient,
				logger:        logrus.NewEntry(logrus.New()),
			},
			want: nil,
		},
		{
			name: "error case - nil dbAdapter",
			args: args{
				tzktAdapter:   providedTZKTAPI,
				dbAdapter:     nil,
				metricsClient: providedMetricsClient,
				logger:        logrus.NewEntry(logrus.New()),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSyncDelegationsFunc(tt.args.tzktAdapter, tt.args.dbAdapter, tt.args.metricsClient, tt.args.logger)
			if (got == nil) != (tt.want == nil) {
				t.Errorf("NewSyncDelegationsFunc() = %v, want %v", got != nil, tt.want != nil)
			}
		})
	}
}

func Test_syncDelegations_SyncDelegations(t *testing.T) {
	type fields struct {
		dbAdapter      database.Adapter
		logger         *logrus.Entry
		tzktApiAdapter tzktapi.Adapter
	}
	tests := []struct {
		name    string
		fields  fields
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetHighestBlockLevel", mock.Anything).Return(int64(100), nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("FetchDelegationsFromLevel", mock.Anything, int64(100)).
						Return(model.TzktDelegationResponse{}, nil)
					return tzkt
				}(),
			},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "error case - nil context",
			fields: fields{
				dbAdapter: func() database.Adapter {
					m := databasemock.New()
					m.On("GetHighestBlockLevel", mock.AnythingOfType("*context.timerCtx")).Return(int64(100), nil)
					return m
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("FetchDelegationsFromLevel", mock.AnythingOfType("*context.timerCtx"), int64(100)).
						Return(model.TzktDelegationResponse{}, nil)
					return tzkt
				}(),
			},
			ctx:     nil,
			wantErr: false,
		},
		{
			name: "error case - dbAdapter returns error",
			fields: fields{
				dbAdapter: func() database.Adapter {
					m := databasemock.New()
					m.On("GetHighestBlockLevel", mock.Anything).
						Return(int64(0), fmt.Errorf("db error"))
					return m
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegations", mock.Anything, batchSize, 0).
						Return(model.TzktDelegationResponse{}, nil)
					return m
				}(),
			},
			ctx:     context.Background(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &syncDelegations{
				dbAdapter:      tt.fields.dbAdapter,
				logger:         tt.fields.logger,
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			if err := uc.SyncDelegations(tt.ctx); (err != nil) != tt.wantErr {
				t.Errorf("SyncDelegations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_syncDelegations_syncHistoricalDelegations(t *testing.T) {
	type fields struct {
		dbAdapter      database.Adapter
		logger         *logrus.Entry
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("SaveDelegations", mock.Anything, mock.Anything).
						Return(nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("FetchDelegations", mock.Anything, batchSize, 0).
						Return(model.TzktDelegationResponse{
							{
								Status:    "applied",
								Level:     100,
								Timestamp: time.Now(),
								Sender:    model.TzktAddress{Address: "tz1sender"},
								Delegate:  model.TzktDelegate{Address: "tz1delegate"},
								Amount:    1000000,
							},
						}, nil)
					tzkt.On("FetchDelegations", mock.Anything, batchSize, batchSize).
						Return(model.TzktDelegationResponse{}, nil)
					return tzkt
				}(),
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
		{
			name: "error case - context cancelled",
			fields: fields{
				dbAdapter:      databasemock.New(),
				logger:         logrus.NewEntry(logrus.New()),
				tzktApiAdapter: tzktapimock.New(),
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
			},
			wantErr: true,
		},
		{
			name: "error case - tzktApiAdapter returns error",
			fields: fields{
				dbAdapter: databasemock.New(),
				logger:    logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegations", mock.Anything, batchSize, 0).
						Return(model.TzktDelegationResponse{}, fmt.Errorf("api error"))
					return m
				}(),
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &syncDelegations{
				dbAdapter:      tt.fields.dbAdapter,
				logger:         tt.fields.logger,
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			if err := uc.syncHistoricalDelegations(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("syncHistoricalDelegations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_syncDelegations_syncIncrementalDelegations(t *testing.T) {
	type fields struct {
		dbAdapter      database.Adapter
		logger         *logrus.Entry
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		ctx   context.Context
		level int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("SaveDelegations", mock.Anything, mock.Anything).Return(nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("FetchDelegationsFromLevel", mock.Anything, int64(100)).
						Return(model.TzktDelegationResponse{
							{
								Status:    "applied",
								Level:     101,
								Timestamp: time.Now(),
								Sender:    model.TzktAddress{Address: "tz1sender"},
								Delegate:  model.TzktDelegate{Address: "tz1delegate"},
								Amount:    1000000,
							},
						}, nil)
					return tzkt
				}(),
			},
			args: args{
				ctx:   context.Background(),
				level: 100,
			},
			wantErr: false,
		},
		{
			name: "error case - context cancelled",
			fields: fields{
				dbAdapter: databasemock.New(),
				logger:    logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegationsFromLevel", mock.MatchedBy(func(ctx context.Context) bool {
						select {
						case <-ctx.Done():
							return true
						default:
							return false
						}
					}), int64(100)).Return(model.TzktDelegationResponse{}, context.Canceled)
					return m
				}(),
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				level: 100,
			},
			wantErr: true,
		},
		{
			name: "error case - tzktApiAdapter returns error",
			fields: fields{
				dbAdapter: databasemock.New(),
				logger:    logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					m := tzktapimock.New()
					m.On("FetchDelegationsFromLevel", mock.Anything, int64(100)).
						Return(model.TzktDelegationResponse{}, fmt.Errorf("api error"))
					return m
				}(),
			},
			args: args{
				ctx:   context.Background(),
				level: 100,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &syncDelegations{
				dbAdapter:      tt.fields.dbAdapter,
				logger:         tt.fields.logger,
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			if err := uc.syncIncrementalDelegations(tt.args.ctx, tt.args.level); (err != nil) != tt.wantErr {
				t.Errorf("syncIncrementalDelegations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_syncDelegations_withMonitorer(t *testing.T) {
	type fields struct {
		dbAdapter      database.Adapter
		logger         *logrus.Entry
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		syncDelegations SyncDelegationsFunc
		metricsClient   metrics.Adapter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   SyncDelegationsFunc
	}{
		{
			name: "nominal case",
			fields: fields{
				dbAdapter:      databasemock.New(),
				logger:         logrus.NewEntry(logrus.New()),
				tzktApiAdapter: tzktapimock.New(),
			},
			args: args{
				syncDelegations: func(ctx context.Context) error { return nil },
				metricsClient:   metricsnoop.New(),
			},
			want: func(ctx context.Context) error { return nil },
		},
		{
			name: "error case - syncDelegations returns error",
			fields: fields{
				dbAdapter:      databasemock.New(),
				logger:         logrus.NewEntry(logrus.New()),
				tzktApiAdapter: tzktapimock.New(),
			},
			args: args{
				syncDelegations: func(ctx context.Context) error { return fmt.Errorf("sync error") },
				metricsClient:   metricsnoop.New(),
			},
			want: func(ctx context.Context) error { return fmt.Errorf("sync error") },
		},
		{
			name: "error case - metricsClient is nil",
			fields: fields{
				dbAdapter:      databasemock.New(),
				logger:         logrus.NewEntry(logrus.New()),
				tzktApiAdapter: tzktapimock.New(),
			},
			args: args{
				syncDelegations: func(ctx context.Context) error { return nil },
				metricsClient:   nil,
			},
			want: func(ctx context.Context) error { return nil },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &syncDelegations{
				dbAdapter:      tt.fields.dbAdapter,
				logger:         tt.fields.logger,
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			got := uc.withMonitorer(tt.args.syncDelegations, tt.args.metricsClient)
			err := got(context.Background())
			expectedErr := tt.args.syncDelegations(context.Background())
			if (err != nil) != (expectedErr != nil) {
				t.Errorf("withMonitorer() error = %v, expected error = %v", err, expectedErr)
			}
		})
	}
}
