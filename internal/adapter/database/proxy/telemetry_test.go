package proxy

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/adapter/database"
	databasemock "github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	metricsmemory "github.com/tezos-delegation-service/internal/adapter/metrics/impl/memory"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	"github.com/tezos-delegation-service/internal/model"
)

func Test_New(t *testing.T) {
	providedMock := databasemock.New()
	providedMetrics := metricsnoop.New()

	type args struct {
		db       database.Adapter
		implType string
		metrics  metrics.Adapter
	}
	tests := []struct {
		name string
		args args
		want database.Adapter
	}{
		{
			name: "nominal case",
			args: args{
				db:       providedMock,
				metrics:  providedMetrics,
				implType: "api",
			},
			want: &TelemetryWrapper{
				metrics:  providedMetrics,
				db:       providedMock,
				implType: "api",
			},
		},
		{
			name: "error case - nil db",
			args: args{
				db:      nil,
				metrics: providedMetrics,
			},
			want: nil,
		},
		{
			name: "error case - nil metrics",
			args: args{
				db:       providedMock,
				metrics:  nil,
				implType: "api",
			},
			want: &TelemetryWrapper{
				metrics:  nil,
				db:       providedMock,
				implType: "api",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.db, tt.args.implType, tt.args.metrics); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TelemetryWrapper_Close(t *testing.T) {
	type fields struct {
		metrics  metrics.Adapter
		db       database.Adapter
		implType string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("Close").Return(nil)
					return m
				}(),
				implType: "api",
			},
			wantErr: false,
		},
		{
			name: "error case - db error",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("Close").Return(errors.New("boom"))
					return m
				}(),
				implType: "api",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				metrics:  tt.fields.metrics,
				db:       tt.fields.db,
				implType: tt.fields.implType,
			}
			if err := w.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_TelemetryWrapper_CountDelegations(t *testing.T) {
	type fields struct {
		metrics  metrics.Adapter
		db       database.Adapter
		implType string
	}
	type args struct {
		ctx  context.Context
		year uint16
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("CountDelegations", mock.Anything, uint16(2025)).
						Return(10, nil)
					return m
				}(),
				implType: "api",
			},
			args: args{
				ctx:  context.TODO(),
				year: 2025,
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "error case - db error",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("CountDelegations", mock.Anything, uint16(2025)).
						Return(0, errors.New("db error"))
					return m
				}(),
				implType: "api",
			},
			args: args{
				ctx:  context.TODO(),
				year: 2025,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				metrics:  tt.fields.metrics,
				db:       tt.fields.db,
				implType: tt.fields.implType,
			}
			got, err := w.CountDelegations(tt.args.ctx, tt.args.year)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountDelegations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CountDelegations() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TelemetryWrapper_GetDelegations(t *testing.T) {
	type fields struct {
		metrics  metrics.Adapter
		db       database.Adapter
		implType string
	}
	type args struct {
		ctx             context.Context
		page            uint32
		limit           uint16
		year            uint16
		maxDelegationID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.Delegation
		want1   int
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("GetDelegations", mock.Anything, uint32(1), uint16(10), uint16(2025), uint64(0)).
						Return([]model.Delegation{
							{Amount: 100},
							{Amount: 200},
						}, 2, nil)
					return m
				}(),
				implType: "api",
			},
			args: args{
				ctx:   context.TODO(),
				page:  1,
				limit: 10,
				year:  2025,
			},
			want: []model.Delegation{
				{Amount: 100},
				{Amount: 200},
			},
			want1:   2,
			wantErr: false,
		},
		{
			name: "error case - db error",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("GetDelegations", mock.Anything, uint32(1), uint16(10), uint16(2025), uint64(0)).
						Return([]model.Delegation{}, 0, errors.New("db error"))
					return m
				}(),
				implType: "api",
			},
			args: args{
				ctx:   context.TODO(),
				page:  1,
				limit: 10,
				year:  2025,
			},
			want:    []model.Delegation{},
			want1:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				metrics:  tt.fields.metrics,
				db:       tt.fields.db,
				implType: tt.fields.implType,
			}
			got, got1, err := w.GetDelegations(tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year, tt.args.maxDelegationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDelegations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDelegations() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetDelegations() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_TelemetryWrapper_GetHighestBlockLevel(t *testing.T) {
	type fields struct {
		metrics  metrics.Adapter
		db       database.Adapter
		implType string
	}
	tests := []struct {
		name    string
		fields  fields
		ctx     context.Context
		want    uint64
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("GetHighestBlockLevel", mock.Anything).
						Return(uint64(1000), nil)
					return m
				}(),
				implType: "api",
			},
			ctx:     context.TODO(),
			want:    1000,
			wantErr: false,
		},
		{
			name: "error case - db error",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("GetHighestBlockLevel", mock.Anything).
						Return(uint64(0), errors.New("db error"))
					return m
				}(),
				implType: "api",
			},
			ctx:     context.TODO(),
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				metrics:  tt.fields.metrics,
				db:       tt.fields.db,
				implType: tt.fields.implType,
			}
			got, err := w.GetHighestBlockLevel(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHighestBlockLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetHighestBlockLevel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TelemetryWrapper_GetLatestDelegation(t *testing.T) {
	type fields struct {
		metrics  metrics.Adapter
		db       database.Adapter
		implType string
	}
	tests := []struct {
		name    string
		fields  fields
		ctx     context.Context
		want    *model.Delegation
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("GetLatestDelegation", mock.Anything).
						Return(&model.Delegation{Amount: 100}, nil)
					return m
				}(),
				implType: "api",
			},
			ctx:     context.TODO(),
			want:    &model.Delegation{Amount: 100},
			wantErr: false,
		},
		{
			name: "error case - db error",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("GetLatestDelegation", mock.Anything).
						Return((*model.Delegation)(nil), errors.New("db error"))
					return m
				}(),
				implType: "api",
			},
			ctx:     context.TODO(),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				metrics:  tt.fields.metrics,
				db:       tt.fields.db,
				implType: tt.fields.implType,
			}
			got, err := w.GetLatestDelegation(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestDelegation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLatestDelegation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TelemetryWrapper_Ping(t *testing.T) {
	type fields struct {
		metrics  metrics.Adapter
		db       database.Adapter
		implType string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "nominal case",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("Ping").Return(nil)
					return m
				}(),
				implType: "api",
			},
			wantErr: false,
		},
		{
			name: "error case - db error",
			fields: fields{
				metrics: metricsmemory.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("Ping").Return(errors.New("db error"))
					return m
				}(),
				implType: "api",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				metrics:  tt.fields.metrics,
				db:       tt.fields.db,
				implType: tt.fields.implType,
			}
			if err := w.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_TelemetryWrapper_SaveDelegation(t *testing.T) {
	type fields struct {
		metrics  metrics.Adapter
		db       database.Adapter
		implType string
	}
	type args struct {
		ctx        context.Context
		delegation *model.Delegation
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
				metrics: metricsnoop.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("SaveDelegation", mock.Anything, mock.Anything).
						Return(nil)
					return m
				}(),
				implType: "api",
			},
			args: args{
				ctx:        context.TODO(),
				delegation: &model.Delegation{Amount: 100},
			},
			wantErr: false,
		},
		{
			name: "error case - db error",
			fields: fields{
				metrics: metricsnoop.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("SaveDelegation", mock.Anything, mock.Anything).
						Return(errors.New("db error"))
					return m
				}(),
				implType: "api",
			},
			args: args{
				ctx:        context.TODO(),
				delegation: &model.Delegation{Amount: 100},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				metrics:  tt.fields.metrics,
				db:       tt.fields.db,
				implType: "api",
			}
			if err := w.SaveDelegation(tt.args.ctx, tt.args.delegation); (err != nil) != tt.wantErr {
				t.Errorf("SaveDelegation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_TelemetryWrapper_SaveDelegations(t *testing.T) {
	type fields struct {
		metrics  metrics.Adapter
		db       database.Adapter
		implType string
	}
	type args struct {
		ctx         context.Context
		delegations []*model.Delegation
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
				metrics: metricsnoop.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("SaveDelegations", mock.Anything, mock.Anything).Return(nil)
					return m
				}(),
				implType: "api",
			},
			args: args{
				ctx: context.TODO(),
				delegations: []*model.Delegation{
					{Amount: 100},
					{Amount: 200},
				},
			},
			wantErr: false,
		},
		{
			name: "error case - repo error",
			fields: fields{
				metrics: metricsnoop.New(),
				db: func() database.Adapter {
					m := databasemock.New()
					m.On("SaveDelegations", mock.Anything, mock.Anything).
						Return(errors.New("repo error"))
					return m
				}(),
			},
			args: args{
				ctx: context.TODO(),
				delegations: []*model.Delegation{
					{Amount: 100},
					{Amount: 200},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &TelemetryWrapper{
				metrics:  tt.fields.metrics,
				db:       tt.fields.db,
				implType: "api",
			}
			if err := w.SaveDelegations(tt.args.ctx, tt.args.delegations); (err != nil) != tt.wantErr {
				t.Errorf("SaveDelegations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
