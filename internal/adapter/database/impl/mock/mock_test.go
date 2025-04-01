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

func Test_Mock_CountDelegations(t *testing.T) {
	type args struct {
		ctx  context.Context
		year int
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("CountDelegations", mock.Anything, 2023).
					Return(10, nil)
				return m
			}(),
			args: args{
				ctx:  context.Background(),
				year: 2023,
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("CountDelegations", mock.Anything, 2023).
					Return(0, errors.New("count error"))
				return m
			}(),
			args: args{
				ctx:  context.Background(),
				year: 2023,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			got, err := m.CountDelegations(tt.args.ctx, tt.args.year)
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

func Test_Mock_GetDelegations(t *testing.T) {
	type args struct {
		ctx   context.Context
		page  int
		limit int
		year  int
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		want    []model.Delegation
		want1   int
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("GetDelegations", mock.Anything, 1, 10, 2023).
					Return([]model.Delegation{{ID: 1}}, 1, nil)
				return m
			}(),
			args: args{
				ctx:   context.Background(),
				page:  1,
				limit: 10,
				year:  2023,
			},
			want:    []model.Delegation{{ID: 1}},
			want1:   1,
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("GetDelegations", mock.Anything, 1, 10, 2023).
					Return(nil, 0, errors.New("get delegations error"))
				return m
			}(),
			args: args{
				ctx:   context.Background(),
				page:  1,
				limit: 10,
				year:  2023,
			},
			want:    nil,
			want1:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			got, got1, err := m.GetDelegations(tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year)
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

func Test_Mock_GetHighestBlockLevel(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "nominal case",
			mock: func() *Mock {
				m := New()
				m.On("GetHighestBlockLevel", mock.Anything).
					Return(int64(100), nil)
				return m
			}(),
			args: args{
				ctx: context.Background(),
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "error case",
			mock: func() *Mock {
				m := New()
				m.On("GetHighestBlockLevel", mock.Anything).
					Return(int64(0), errors.New("get highest block level error"))
				return m
			}(),
			args: args{
				ctx: context.Background(),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			got, err := m.GetHighestBlockLevel(tt.args.ctx)
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
