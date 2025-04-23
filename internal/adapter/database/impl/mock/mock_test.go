package mock

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/model"
)

func Test_Mock_Close(t *testing.T) {
	tests := []struct {
		name    string
		mock    *Mock
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("Close").
					Return(nil)
				return m
			}(),
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("Close").
					Return(errors.New("close error"))
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			if err := m.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Mock_GetDelegations(t *testing.T) {
	type args struct {
		ctx             context.Context
		page            uint32
		limit           uint16
		year            uint16
		maxDelegationID uint64
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		want    []model.Delegation
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("GetDelegations", mock.Anything, uint32(1), uint16(10), uint16(2025), uint64(0)).
					Return([]model.Delegation{{ID: 1}}, nil)
				return m
			}(),
			args: args{
				ctx:             context.Background(),
				page:            1,
				limit:           10,
				year:            2025,
				maxDelegationID: 0,
			},
			want:    []model.Delegation{{ID: 1}},
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("GetDelegations", mock.Anything, uint32(1), uint16(10), uint16(2025), uint64(0)).
					Return([]model.Delegation(nil), errors.New("get delegations error"))
				return m
			}(),
			args: args{
				ctx:             context.Background(),
				page:            1,
				limit:           10,
				year:            2025,
				maxDelegationID: 0,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			got, err := m.GetDelegations(tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year, tt.args.maxDelegationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDelegations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDelegations() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Mock_GetHighestBlockLevel(t *testing.T) {
	tests := []struct {
		name    string
		mock    *Mock
		want    uint64
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("GetHighestBlockLevel", mock.Anything).
					Return(uint64(100), nil)
				return m
			}(),
			want:    100,
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("GetHighestBlockLevel", mock.Anything).
					Return(uint64(0), errors.New("get highest block level error"))
				return m
			}(),
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			got, err := m.GetHighestBlockLevel(context.Background())
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

func Test_Mock_GetLatestDelegation(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		want    *model.Delegation
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("GetLatestDelegation", mock.Anything).
					Return(&model.Delegation{ID: 1}, nil)
				return m
			}(),
			args: args{
				ctx: context.Background(),
			},
			want:    &model.Delegation{ID: 1},
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("GetLatestDelegation", mock.Anything).
					Return(nil, errors.New("get latest delegation error"))
				return m
			}(),
			args: args{
				ctx: context.Background(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			got, err := m.GetLatestDelegation(tt.args.ctx)
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

func Test_Mock_Ping(t *testing.T) {
	tests := []struct {
		name    string
		mock    *Mock
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("Ping").Return(nil)
				return m
			}(),
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("Ping").Return(errors.New("ping error"))
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			if err := m.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Mock_SaveDelegation(t *testing.T) {
	type args struct {
		ctx        context.Context
		delegation *model.Delegation
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("SaveDelegation", mock.Anything, &model.Delegation{ID: 1}).
					Return(nil)
				return m
			}(),
			args: args{
				ctx:        context.Background(),
				delegation: &model.Delegation{ID: 1},
			},
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("SaveDelegation", mock.Anything, &model.Delegation{ID: 1}).
					Return(errors.New("save delegation error"))
				return m
			}(),
			args: args{
				ctx:        context.Background(),
				delegation: &model.Delegation{ID: 1},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			if err := m.SaveDelegation(tt.args.ctx, tt.args.delegation); (err != nil) != tt.wantErr {
				t.Errorf("SaveDelegation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Mock_SaveDelegations(t *testing.T) {
	type args struct {
		ctx         context.Context
		delegations []*model.Delegation
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("SaveDelegations", mock.Anything, []*model.Delegation{{ID: 1}}).
					Return(nil)
				return m
			}(),
			args: args{
				ctx:         context.Background(),
				delegations: []*model.Delegation{{ID: 1}},
			},
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("SaveDelegations", mock.Anything, []*model.Delegation{{ID: 1}}).
					Return(errors.New("save delegations error"))
				return m
			}(),
			args: args{
				ctx:         context.Background(),
				delegations: []*model.Delegation{{ID: 1}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			if err := m.SaveDelegations(tt.args.ctx, tt.args.delegations); (err != nil) != tt.wantErr {
				t.Errorf("SaveDelegations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_New(t *testing.T) {
	providedMock := &Mock{}
	t.Run("nominal", func(t *testing.T) {
		if got := New(); !reflect.DeepEqual(got, providedMock) {
			t.Errorf("New() = %v, want %v", got, providedMock)
		}
	})
}
