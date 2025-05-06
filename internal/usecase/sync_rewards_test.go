package usecase

import (
	"context"
	"errors"
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

func Test_NewSyncRewardsFunc(t *testing.T) {
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
		want bool // On vérifie simplement si la fonction retourne non-nil
	}{
		{
			name: "nominal case",
			args: args{
				tzktAdapter:   providedTZKTAPI,
				dbAdapter:     mockDbAdapter,
				metricsClient: providedMetricsClient,
				logger:        logrus.NewEntry(logrus.New()),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSyncRewardsFunc(tt.args.tzktAdapter, tt.args.dbAdapter, tt.args.metricsClient, tt.args.logger)
			if (got != nil) != tt.want {
				t.Errorf("NewSyncRewardsFunc() = %v, want %v", got != nil, tt.want)
			}
		})
	}
}

func Test_SyncRewards_Sync(t *testing.T) {
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
			name: "nominal case - up to date",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetLastSyncedRewardCycle", mock.Anything).
						Return(10, nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.Anything).
						Return(10, nil)
					return tzkt
				}(),
			},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "nominal case - needs sync",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetLastSyncedRewardCycle", mock.Anything).
						Return(9, nil)
					db.On("GetActiveDelegators", mock.Anything).
						Return([]model.WalletAddress{"tz1delegator1", "tz1delegator2"}, nil)
					db.On("GetBakerForDelegatorAtCycle", mock.Anything, model.WalletAddress("tz1delegator1"), 10).
						Return(model.WalletAddress("tz1baker1"), nil)
					db.On("GetBakerForDelegatorAtCycle", mock.Anything, model.WalletAddress("tz1delegator2"), 10).
						Return(model.WalletAddress("tz1baker2"), nil)
					db.On("SaveRewards", mock.Anything, mock.Anything).
						Return(nil)
					db.On("SaveLastSyncedRewardCycle", mock.Anything, 10).
						Return(nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.Anything).
						Return(10, nil)
					tzkt.On("FetchRewardsForCycle", mock.Anything, model.WalletAddress("tz1delegator1"), model.WalletAddress("tz1baker1"), 10).
						Return([]model.Reward{
							{
								RecipientAddress: "tz1delegator1",
								SourceAddress:    "tz1baker1",
								Cycle:            10,
								Amount:           5.5,
								Timestamp:        time.Now().Unix(),
							},
						}, nil)
					tzkt.On("FetchRewardsForCycle", mock.Anything, model.WalletAddress("tz1delegator2"), model.WalletAddress("tz1baker2"), 10).
						Return([]model.Reward{
							{
								RecipientAddress: "tz1delegator2",
								SourceAddress:    "tz1baker2",
								Cycle:            10,
								Amount:           3.3,
								Timestamp:        time.Now().Unix(),
							},
						}, nil)
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
					db := databasemock.New()
					db.On("GetLastSyncedRewardCycle", mock.AnythingOfType("*context.timerCtx")).
						Return(9, nil)
					db.On("GetActiveDelegators", mock.AnythingOfType("*context.timerCtx")).
						Return([]model.WalletAddress{}, nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.AnythingOfType("*context.timerCtx")).
						Return(10, nil)
					return tzkt
				}(),
			},
			ctx:     nil,
			wantErr: false,
		},
		{
			name: "error case - GetCurrentCycle error",
			fields: fields{
				dbAdapter: databasemock.New(),
				logger:    logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.Anything).
						Return(0, errors.New("api error"))
					return tzkt
				}(),
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "error case - GetLastSyncedRewardCycle error",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetLastSyncedRewardCycle", mock.Anything).
						Return(0, errors.New("db error"))
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.Anything).
						Return(10, nil)
					return tzkt
				}(),
			},
			ctx:     context.Background(),
			wantErr: false, // No error because the function handles the db error and continues
		},
		{
			name: "error case - GetActiveDelegators error",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetLastSyncedRewardCycle", mock.Anything).
						Return(9, nil)
					db.On("GetActiveDelegators", mock.Anything).
						Return([]model.WalletAddress{}, errors.New("delegators error"))
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.Anything).
						Return(10, nil)
					return tzkt
				}(),
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "error case - context cancelled",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetLastSyncedRewardCycle", mock.Anything).
						Return(9, nil)
					db.On("GetActiveDelegators", mock.Anything).
						Return([]model.WalletAddress{"tz1delegator1"}, nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.Anything).
						Return(10, nil)
					return tzkt
				}(),
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			}(),
			wantErr: true,
		},
		{
			name: "error case - SaveRewards error",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetLastSyncedRewardCycle", mock.Anything).
						Return(9, nil)
					db.On("GetActiveDelegators", mock.Anything).
						Return([]model.WalletAddress{"tz1delegator1"}, nil)
					db.On("GetBakerForDelegatorAtCycle", mock.Anything, model.WalletAddress("tz1delegator1"), 10).
						Return(model.WalletAddress("tz1baker1"), nil)
					db.On("SaveRewards", mock.Anything, mock.Anything).
						Return(errors.New("save error"))
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.Anything).
						Return(10, nil)
					tzkt.On("FetchRewardsForCycle", mock.Anything, model.WalletAddress("tz1delegator1"), model.WalletAddress("tz1baker1"), 10).
						Return([]model.Reward{
							{
								RecipientAddress: "tz1delegator1",
								SourceAddress:    "tz1baker1",
								Cycle:            10,
								Amount:           5.5,
								Timestamp:        time.Now().Unix(),
							},
						}, nil)
					return tzkt
				}(),
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "error case - SaveLastSyncedRewardCycle error",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetLastSyncedRewardCycle", mock.Anything).
						Return(9, nil)
					db.On("GetActiveDelegators", mock.Anything).
						Return([]model.WalletAddress{"tz1delegator1"}, nil)
					db.On("GetBakerForDelegatorAtCycle", mock.Anything, model.WalletAddress("tz1delegator1"), 10).
						Return(model.WalletAddress("tz1baker1"), nil)
					db.On("SaveRewards", mock.Anything, mock.Anything).
						Return(nil)
					db.On("SaveLastSyncedRewardCycle", mock.Anything, 10).
						Return(errors.New("sync save error"))
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
				tzktApiAdapter: func() tzktapi.Adapter {
					tzkt := tzktapimock.New()
					tzkt.On("GetCurrentCycle", mock.Anything).
						Return(10, nil)
					tzkt.On("FetchRewardsForCycle", mock.Anything, model.WalletAddress("tz1delegator1"), model.WalletAddress("tz1baker1"), 10).
						Return([]model.Reward{
							{
								RecipientAddress: "tz1delegator1",
								SourceAddress:    "tz1baker1",
								Cycle:            10,
								Amount:           5.5,
								Timestamp:        time.Now().Unix(),
							},
						}, nil)
					return tzkt
				}(),
			},
			ctx:     context.Background(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &SyncRewards{
				batchSize:      1000,
				dbAdapter:      tt.fields.dbAdapter,
				logger:         tt.fields.logger,
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			if err := uc.Sync(tt.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_SyncRewards_saveRewardsBatch(t *testing.T) {
	type fields struct {
		batchSize      int
		dbAdapter      database.Adapter
		logger         *logrus.Entry
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		ctx     context.Context
		rewards []model.Reward
		cycle   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nominal case - empty rewards",
			fields: fields{
				batchSize: 1000,
				dbAdapter: databasemock.New(),
				logger:    logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:     context.Background(),
				rewards: []model.Reward{},
				cycle:   10,
			},
			wantErr: false,
		},
		{
			name: "nominal case - with rewards",
			fields: fields{
				batchSize: 2,
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("SaveRewards", mock.Anything, mock.Anything).
						Return(nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx: context.Background(),
				rewards: []model.Reward{
					{
						RecipientAddress: "tz1delegator1",
						SourceAddress:    "tz1baker1",
						Cycle:            10,
						Amount:           5.5,
						Timestamp:        time.Now().Unix(),
					},
					{
						RecipientAddress: "tz1delegator2",
						SourceAddress:    "tz1baker2",
						Cycle:            10,
						Amount:           3.3,
						Timestamp:        time.Now().Unix(),
					},
					{
						RecipientAddress: "tz1delegator3",
						SourceAddress:    "tz1baker3",
						Cycle:            10,
						Amount:           7.7,
						Timestamp:        time.Now().Unix(),
					},
				},
				cycle: 10,
			},
			wantErr: false,
		},
		{
			name: "error case - SaveRewards error",
			fields: fields{
				batchSize: 1000,
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("SaveRewards", mock.Anything, mock.Anything).
						Return(fmt.Errorf("save rewards error"))
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx: context.Background(),
				rewards: []model.Reward{
					{
						RecipientAddress: "tz1delegator1",
						SourceAddress:    "tz1baker1",
						Cycle:            10,
						Amount:           5.5,
						Timestamp:        time.Now().Unix(),
					},
				},
				cycle: 10,
			},
			wantErr: true,
		},
		{
			name: "error case - context cancelled",
			fields: fields{
				batchSize: 1,
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("SaveRewards", mock.Anything, mock.Anything).
						Return(nil)
					return db
				}(),
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel() // Cancel immediately
					return ctx
				}(),
				rewards: []model.Reward{
					{
						RecipientAddress: "tz1delegator1",
						SourceAddress:    "tz1baker1",
						Cycle:            10,
						Amount:           5.5,
						Timestamp:        time.Now().Unix(),
					},
					{
						RecipientAddress: "tz1delegator2",
						SourceAddress:    "tz1baker2",
						Cycle:            10,
						Amount:           3.3,
						Timestamp:        time.Now().Unix(),
					},
				},
				cycle: 10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &SyncRewards{
				batchSize:      tt.fields.batchSize,
				dbAdapter:      tt.fields.dbAdapter,
				logger:         tt.fields.logger,
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			if err := uc.saveRewardsBatch(tt.args.ctx, tt.args.rewards, tt.args.cycle); (err != nil) != tt.wantErr {
				t.Errorf("saveRewardsBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_SyncRewards_withMonitorer(t *testing.T) {
	type fields struct {
		batchSize      int
		dbAdapter      database.Adapter
		logger         *logrus.Entry
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		syncRewards  model.SyncFunc
		metricsClient metrics.Adapter
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
				batchSize:      1000,
				dbAdapter:      databasemock.New(),
				logger:         logrus.NewEntry(logrus.New()),
				tzktApiAdapter: tzktapimock.New(),
			},
			args: args{
				syncRewards:   func(ctx context.Context) error { return nil },
				metricsClient: metricsnoop.New(),
			},
			wantErr: false,
		},
		{
			name: "error case - syncRewards returns error",
			fields: fields{
				batchSize:      1000,
				dbAdapter:      databasemock.New(),
				logger:         logrus.NewEntry(logrus.New()),
				tzktApiAdapter: tzktapimock.New(),
			},
			args: args{
				syncRewards:   func(ctx context.Context) error { return fmt.Errorf("sync error") },
				metricsClient: metricsnoop.New(),
			},
			wantErr: true,
		},
		{
			name: "nominal case - nil metricsClient",
			fields: fields{
				batchSize:      1000,
				dbAdapter:      databasemock.New(),
				logger:         logrus.NewEntry(logrus.New()),
				tzktApiAdapter: tzktapimock.New(),
			},
			args: args{
				syncRewards:   func(ctx context.Context) error { return nil },
				metricsClient: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &SyncRewards{
				batchSize:      tt.fields.batchSize,
				dbAdapter:      tt.fields.dbAdapter,
				logger:         tt.fields.logger,
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			monitoredFunc := uc.withMonitorer(tt.args.syncRewards, tt.args.metricsClient)
			if err := monitoredFunc(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("withMonitorer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}