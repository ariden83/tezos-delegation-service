package poller

import (
	"context"
	"errors"
	"reflect"
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
				tzktAdapter:     tzktapimock.New(),
				dbAdapter:       databasemock.New(),
				pollingInterval: time.Second,
				metricClient:    metrisnoop.New(),
				logger:          logrus.NewEntry(logrus.New()),
			},
			want: &Poller{
				dbAdapter:       databasemock.New(),
				logger:          logrus.NewEntry(logrus.New()).WithField("component", "poller"),
				pollingInterval: time.Second,
				tzktAdapter:     tzktapimock.New(),
				ucSyncDelegations: usecase.NewSyncDelegationsFunc(tzktapimock.New(), databasemock.New(), metrisnoop.New(),
					logrus.NewEntry(logrus.New())),
			},
		},
		{
			name: "Error case - nil tzktAdapter",
			args: args{
				tzktAdapter:     nil,
				dbAdapter:       databasemock.New(),
				pollingInterval: time.Second,
				metricClient:    metrisnoop.New(),
				logger:          logrus.NewEntry(logrus.New()),
			},
			want: nil,
		},
		{
			name: "Error case - nil dbAdapter",
			args: args{
				tzktAdapter:     tzktapimock.New(),
				dbAdapter:       nil,
				pollingInterval: time.Second,
				metricClient:    metrisnoop.New(),
				logger:          logrus.NewEntry(logrus.New()),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.tzktAdapter,
				tt.args.dbAdapter,
				tt.args.pollingInterval,
				tt.args.metricClient,
				tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Poller_Run(t *testing.T) {
	type fields struct {
		dbAdapter         database.Adapter
		logger            *logrus.Entry
		pollingInterval   time.Duration
		pollingCtx        context.Context
		pollingCancel     context.CancelFunc
		pollingWg         *sync.WaitGroup
		tzktAdapter       tzktapi.Adapter
		ucSyncDelegations usecase.SyncDelegationsFunc
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Nominal case",
			fields: fields{
				dbAdapter:       databasemock.New(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   nil,
				pollingWg:       &sync.WaitGroup{},
				tzktAdapter:     tzktapimock.New(),
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
						Return(0, errors.New("db error"))
					return m
				}(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   nil,
				pollingWg:       &sync.WaitGroup{},
				tzktAdapter:     tzktapimock.New(),
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
						Return(0, nil)
					return m
				}(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   nil,
				pollingWg:       &sync.WaitGroup{},
				tzktAdapter:     tzktapimock.New(),
				ucSyncDelegations: func(ctx context.Context) error {
					return errors.New("sync error")
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				dbAdapter:         tt.fields.dbAdapter,
				logger:            tt.fields.logger,
				pollingInterval:   tt.fields.pollingInterval,
				pollingCtx:        tt.fields.pollingCtx,
				pollingCancel:     tt.fields.pollingCancel,
				pollingWg:         tt.fields.pollingWg,
				tzktAdapter:       tt.fields.tzktAdapter,
				ucSyncDelegations: tt.fields.ucSyncDelegations,
			}
			p.Run()
		})
	}
}

func Test_Poller_StartPolling(t *testing.T) {
	type fields struct {
		dbAdapter         database.Adapter
		logger            *logrus.Entry
		pollingInterval   time.Duration
		pollingCtx        context.Context
		pollingCancel     context.CancelFunc
		pollingWg         *sync.WaitGroup
		tzktAdapter       tzktapi.Adapter
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
				dbAdapter:       databasemock.New(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   nil,
				pollingWg:       &sync.WaitGroup{},
				tzktAdapter:     tzktapimock.New(),
				ucSyncDelegations: func(ctx context.Context) error {
					return nil
				},
			},
			ctx: context.Background(),
		},
		{
			name: "Error case - pollingWg is nil",
			fields: fields{
				dbAdapter:       databasemock.New(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   nil,
				pollingWg:       nil,
				tzktAdapter:     tzktapimock.New(),
				ucSyncDelegations: func(ctx context.Context) error {
					return nil
				},
			},
			ctx: context.Background(),
		},
		{
			name: "Error case - ucSyncDelegations returns error",
			fields: fields{
				dbAdapter:       databasemock.New(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   nil,
				pollingWg:       &sync.WaitGroup{},
				tzktAdapter:     tzktapimock.New(),
				ucSyncDelegations: func(ctx context.Context) error {
					return errors.New("sync error")
				},
			},
			ctx: context.Background(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				dbAdapter:         tt.fields.dbAdapter,
				logger:            tt.fields.logger,
				pollingInterval:   tt.fields.pollingInterval,
				pollingCtx:        tt.fields.pollingCtx,
				pollingCancel:     tt.fields.pollingCancel,
				pollingWg:         tt.fields.pollingWg,
				tzktAdapter:       tt.fields.tzktAdapter,
				ucSyncDelegations: tt.fields.ucSyncDelegations,
			}
			p.StartPolling(tt.ctx)
		})
	}
}

func Test_Poller_StopPolling(t *testing.T) {
	type fields struct {
		dbAdapter         database.Adapter
		logger            *logrus.Entry
		pollingInterval   time.Duration
		pollingCtx        context.Context
		pollingCancel     context.CancelFunc
		pollingWg         *sync.WaitGroup
		tzktAdapter       tzktapi.Adapter
		ucSyncDelegations usecase.SyncDelegationsFunc
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Nominal case",
			fields: fields{
				dbAdapter:       databasemock.New(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   nil,
				pollingWg:       &sync.WaitGroup{},
				tzktAdapter:     tzktapimock.New(),
			},
		},
		{
			name: "Error case - pollingCancel is nil",
			fields: fields{
				dbAdapter:       databasemock.New(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   nil,
				pollingWg:       &sync.WaitGroup{},
				tzktAdapter:     tzktapimock.New(),
			},
		},
		{
			name: "Error case - pollingWg is nil",
			fields: fields{
				dbAdapter:       databasemock.New(),
				logger:          logrus.NewEntry(logrus.New()),
				pollingInterval: time.Second,
				pollingCtx:      context.Background(),
				pollingCancel:   func() {},
				pollingWg:       nil,
				tzktAdapter:     tzktapimock.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				dbAdapter:         tt.fields.dbAdapter,
				logger:            tt.fields.logger,
				pollingInterval:   tt.fields.pollingInterval,
				pollingCtx:        tt.fields.pollingCtx,
				pollingCancel:     tt.fields.pollingCancel,
				pollingWg:         tt.fields.pollingWg,
				tzktAdapter:       tt.fields.tzktAdapter,
				ucSyncDelegations: tt.fields.ucSyncDelegations,
			}
			p.StopPolling()
		})
	}
}
