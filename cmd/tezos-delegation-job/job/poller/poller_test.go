package poller

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/adapter/database"
	databasemock "github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	metrisnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	tzktapimock "github.com/tezos-delegation-service/internal/adapter/tzktapi/impl/mock"
	"github.com/tezos-delegation-service/internal/usecase"
)

func Test_New(t *testing.T) {
	tzktAdapter := tzktapimock.New()
	dbAdapter := databasemock.New()
	pollingInterval := time.Second
	metricClient := metrisnoop.New()
	logger := logrus.NewEntry(logrus.New())

	type args struct {
		tzktAdapter     tzktapi.Adapter
		dbAdapter       database.Adapter
		pollingInterval time.Duration
		metricClient    metrics.Adapter
		logger          *logrus.Entry
	}
	tests := []struct {
		name string
		args args
		want *Poller
	}{
		{
			name: "Nominal case",
			args: args{
				tzktAdapter:     tzktAdapter,
				dbAdapter:       dbAdapter,
				pollingInterval: pollingInterval,
				metricClient:    metricClient,
				logger:          logger,
			},
			want: &Poller{
				dbAdapter:         dbAdapter,
				logger:            logger.WithField("component", "poller"),
				pollingInterval:   pollingInterval,
				tzktAdapter:       tzktAdapter,
				ucSyncDelegations: usecase.NewSyncDelegationsFunc(tzktAdapter, dbAdapter, metricClient, logger),
			},
		},
		{
			name: "Case with nil tzktAdapter",
			args: args{
				tzktAdapter:     nil,
				dbAdapter:       dbAdapter,
				pollingInterval: pollingInterval,
				metricClient:    metricClient,
				logger:          logger,
			},
			want: &Poller{
				dbAdapter:         dbAdapter,
				logger:            logger.WithField("component", "poller"),
				pollingInterval:   pollingInterval,
				tzktAdapter:       nil,
				ucSyncDelegations: usecase.NewSyncDelegationsFunc(nil, dbAdapter, metricClient, logger),
			},
		},
		{
			name: "Case with nil dbAdapter",
			args: args{
				tzktAdapter:     tzktAdapter,
				dbAdapter:       nil,
				pollingInterval: pollingInterval,
				metricClient:    metricClient,
				logger:          logger,
			},
			want: &Poller{
				dbAdapter:         nil,
				logger:            logger.WithField("component", "poller"),
				pollingInterval:   pollingInterval,
				tzktAdapter:       tzktAdapter,
				ucSyncDelegations: usecase.NewSyncDelegationsFunc(tzktAdapter, nil, metricClient, logger),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.tzktAdapter,
				tt.args.dbAdapter,
				tt.args.pollingInterval,
				tt.args.metricClient,
				tt.args.logger)

			if got.dbAdapter != tt.want.dbAdapter ||
				got.pollingInterval != tt.want.pollingInterval ||
				got.tzktAdapter != tt.want.tzktAdapter {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Poller_Run(t *testing.T) {
	type fields struct {
		dbAdapter         database.Adapter
		pollingCancel     context.CancelFunc
		pollingWg         *sync.WaitGroup
		ucSyncDelegations usecase.SyncDelegationsFunc
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Nominal case with highestLevel 0",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetHighestBlockLevel", mock.Anything).
						Return(uint64(0), nil)
					return db
				}(),
				pollingCancel: nil,
				pollingWg:     &sync.WaitGroup{},
				ucSyncDelegations: func(ctx context.Context) error {
					return nil
				},
			},
		},
		{
			name: "Nominal case with highestLevel over 0",
			fields: fields{
				dbAdapter: func() database.Adapter {
					db := databasemock.New()
					db.On("GetHighestBlockLevel", mock.Anything).
						Return(uint64(100), nil)
					return db
				}(),
				pollingCancel: nil,
				pollingWg:     &sync.WaitGroup{},
				ucSyncDelegations: func(ctx context.Context) error {
					return nil
				},
			},
		},
		{
			name: "Error case - GetHighestBlockLevel returns error",
			fields: fields{
				dbAdapter: func() database.Adapter {
					m := databasemock.New()
					m.On("GetHighestBlockLevel", mock.Anything).
						Return(uint64(0), errors.New("db error"))
					return m
				}(),
				pollingCancel: nil,
				pollingWg:     &sync.WaitGroup{},
				ucSyncDelegations: func(ctx context.Context) error {
					return nil
				},
			},
		},
		{
			name: "Error case - ucSyncDelegations returns error",
			fields: fields{
				dbAdapter: func() database.Adapter {
					m := databasemock.New()
					m.On("GetHighestBlockLevel", mock.Anything).
						Return(uint64(0), nil)
					return m
				}(),
				pollingCancel: nil,
				pollingWg:     &sync.WaitGroup{},
				ucSyncDelegations: func(ctx context.Context) error {
					return errors.New("sync error")
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			p := &Poller{
				dbAdapter:         tt.fields.dbAdapter,
				logger:            logrus.NewEntry(logrus.New()),
				pollingInterval:   time.Second,
				pollingCtx:        ctx,
				pollingCancel:     tt.fields.pollingCancel,
				pollingWg:         tt.fields.pollingWg,
				ucSyncDelegations: tt.fields.ucSyncDelegations,
			}
			p.Run(ctx)
		})
	}
}

func Test_Poller_StartPolling(t *testing.T) {
	type fields struct {
		dbAdapter         database.Adapter
		pollingCancel     context.CancelFunc
		pollingWg         *sync.WaitGroup
		ucSyncDelegations usecase.SyncDelegationsFunc
	}
	tests := []struct {
		name   string
		fields fields
		ctx    context.Context
	}{
		{
			name: "Nominal case",
			fields: fields{
				dbAdapter:     databasemock.New(),
				pollingCancel: nil,
				pollingWg:     &sync.WaitGroup{},
				ucSyncDelegations: func(ctx context.Context) error {
					return nil
				},
			},
			ctx: context.Background(),
		},
		{
			name: "Error case - ucSyncDelegations returns error",
			fields: fields{
				dbAdapter:     databasemock.New(),
				pollingCancel: nil,
				pollingWg:     &sync.WaitGroup{},
				ucSyncDelegations: func(ctx context.Context) error {
					return errors.New("sync error")
				},
			},
			ctx: context.Background(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			p := &Poller{
				dbAdapter:         tt.fields.dbAdapter,
				logger:            logrus.NewEntry(logrus.New()),
				pollingInterval:   time.Second,
				pollingCtx:        ctx,
				pollingCancel:     tt.fields.pollingCancel,
				pollingWg:         tt.fields.pollingWg,
				ucSyncDelegations: tt.fields.ucSyncDelegations,
			}
			p.StartPolling(ctx)
		})
	}
}

func Test_Poller_StopPolling(t *testing.T) {
	type fields struct {
		dbAdapter         database.Adapter
		pollingCancel     context.CancelFunc
		pollingWg         *sync.WaitGroup
		ucSyncDelegations usecase.SyncDelegationsFunc
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Nominal case",
			fields: fields{
				dbAdapter:     databasemock.New(),
				pollingCancel: nil,
				pollingWg:     &sync.WaitGroup{},
			},
		},
		{
			name: "Error case - pollingCancel is nil",
			fields: fields{
				dbAdapter:     databasemock.New(),
				pollingCancel: nil,
				pollingWg:     &sync.WaitGroup{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			p := &Poller{
				dbAdapter:         tt.fields.dbAdapter,
				logger:            logrus.NewEntry(logrus.New()),
				pollingInterval:   time.Second,
				pollingCtx:        ctx,
				pollingCancel:     tt.fields.pollingCancel,
				pollingWg:         tt.fields.pollingWg,
				ucSyncDelegations: tt.fields.ucSyncDelegations,
			}
			p.StopPolling()
		})
	}
}
