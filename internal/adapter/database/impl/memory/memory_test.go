package memory

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/model"
)

func Test_Memory_Close(t *testing.T) {
	type fields struct {
		delegations []*model.Delegation
		nextID      int64
		mu          *sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Nominal case",
			fields:  fields{},
			wantErr: false,
		},
		{
			name: "Error case - non-nil error",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				delegations: tt.fields.delegations,
				nextID:      tt.fields.nextID,
				mu:          tt.fields.mu,
			}
			if err := m.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Memory_CountDelegations(t *testing.T) {
	type fields struct {
		delegations []*model.Delegation
		nextID      int64
		mu          *sync.RWMutex
	}
	type args struct {
		in0  context.Context
		year int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Nominal case - no year filter",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Timestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix()},
					{ID: 2, Timestamp: time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC).Unix()},
				},
				nextID: 3,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0:  context.TODO(),
				year: 0,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "Nominal case - with year filter",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Timestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix()},
					{ID: 2, Timestamp: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC).Unix()},
				},
				nextID: 3,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0:  context.TODO(),
				year: 2022,
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Error case - invalid year",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Timestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix()},
				},
				nextID: 2,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0:  context.TODO(),
				year: -1,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				delegations: tt.fields.delegations,
				nextID:      tt.fields.nextID,
				mu:          tt.fields.mu,
			}
			got, err := m.CountDelegations(tt.args.in0, tt.args.year)
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

func Test_Memory_GetDelegations(t *testing.T) {
	type fields struct {
		delegations []*model.Delegation
		nextID      int64
		mu          *sync.RWMutex
	}
	type args struct {
		in0   context.Context
		page  int
		limit int
		year  int
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
			name: "Nominal case - no year filter",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Timestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix()},
					{ID: 2, Timestamp: time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC).Unix()},
				},
				nextID: 3,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0:   context.TODO(),
				page:  1,
				limit: 10,
				year:  0,
			},
			want: []model.Delegation{
				{ID: 2, Timestamp: time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC).Unix()},
				{ID: 1, Timestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix()},
			},
			want1:   2,
			wantErr: false,
		},
		{
			name: "Error case - invalid year",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Timestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix()},
				},
				nextID: 2,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0:   context.TODO(),
				page:  1,
				limit: 10,
				year:  -1,
			},
			want:    nil,
			want1:   0,
			wantErr: true,
		},
		{
			name: "Error case - nil context",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Timestamp: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix()},
				},
				nextID: 2,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0:   nil,
				page:  1,
				limit: 10,
				year:  2022,
			},
			want:    nil,
			want1:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				delegations: tt.fields.delegations,
				nextID:      tt.fields.nextID,
				mu:          tt.fields.mu,
			}
			got, got1, err := m.GetDelegations(tt.args.in0, tt.args.page, tt.args.limit, tt.args.year)
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

func Test_Memory_GetHighestBlockLevel(t *testing.T) {
	type fields struct {
		delegations []*model.Delegation
		nextID      int64
		mu          *sync.RWMutex
	}
	type args struct {
		in0 context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Level: 1},
					{ID: 2, Level: 3},
					{ID: 3, Level: 2},
				},
				nextID: 4,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0: context.TODO(),
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "Error case - no delegations",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			args: args{
				in0: context.TODO(),
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "Error case - nil context",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Level: 1},
				},
				nextID: 2,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0: nil,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				delegations: tt.fields.delegations,
				nextID:      tt.fields.nextID,
				mu:          tt.fields.mu,
			}
			got, err := m.GetHighestBlockLevel(tt.args.in0)
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

func Test_Memory_GetLatestDelegation(t *testing.T) {
	type fields struct {
		delegations []*model.Delegation
		nextID      int64
		mu          *sync.RWMutex
	}
	type args struct {
		in0 context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Delegation
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Level: 1, Timestamp: time.Now().Unix()},
					{ID: 2, Level: 2, Timestamp: time.Now().Unix()},
				},
				nextID: 3,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0: context.TODO(),
			},
			want: &model.Delegation{
				ID:        2,
				Level:     2,
				Timestamp: time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name: "Error case - no delegations",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			args: args{
				in0: context.TODO(),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Error case - nil context",
			fields: fields{
				delegations: []*model.Delegation{
					{ID: 1, Level: 1, Timestamp: time.Now().Unix()},
				},
				nextID: 2,
				mu:     &sync.RWMutex{},
			},
			args: args{
				in0: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				delegations: tt.fields.delegations,
				nextID:      tt.fields.nextID,
				mu:          tt.fields.mu,
			}
			got, err := m.GetLatestDelegation(tt.args.in0)
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

func Test_Memory_Ping(t *testing.T) {
	type fields struct {
		delegations []*model.Delegation
		nextID      int64
		mu          *sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			wantErr: false,
		},
		{
			name: "Error case - non-nil error",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				delegations: tt.fields.delegations,
				nextID:      tt.fields.nextID,
				mu:          tt.fields.mu,
			}
			if err := m.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Memory_SaveDelegation(t *testing.T) {
	type fields struct {
		delegations []*model.Delegation
		nextID      int64
		mu          *sync.RWMutex
	}
	type args struct {
		in0        context.Context
		delegation *model.Delegation
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			args: args{
				in0: context.TODO(),
				delegation: &model.Delegation{
					Timestamp: time.Now().Unix(),
				},
			},
			wantErr: false,
		},
		{
			name: "Error case - nil context",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			args: args{
				in0:        nil,
				delegation: &model.Delegation{},
			},
			wantErr: true,
		},
		{
			name: "Error case - nil delegation",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			args: args{
				in0:        context.TODO(),
				delegation: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				delegations: tt.fields.delegations,
				nextID:      tt.fields.nextID,
				mu:          tt.fields.mu,
			}
			if err := m.SaveDelegation(tt.args.in0, tt.args.delegation); (err != nil) != tt.wantErr {
				t.Errorf("SaveDelegation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Memory_SaveDelegations(t *testing.T) {
	type fields struct {
		delegations []*model.Delegation
		nextID      int64
		mu          *sync.RWMutex
	}
	type args struct {
		in0         context.Context
		delegations []*model.Delegation
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			args: args{
				in0: context.TODO(),
				delegations: []*model.Delegation{
					{ID: 1, Timestamp: time.Now().Unix()},
					{ID: 2, Timestamp: time.Now().Unix()},
				},
			},
			wantErr: false,
		},
		{
			name: "Error case - nil context",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			args: args{
				in0:         nil,
				delegations: []*model.Delegation{},
			},
			wantErr: true,
		},
		{
			name: "Error case - nil delegations",
			fields: fields{
				delegations: []*model.Delegation{},
				nextID:      1,
				mu:          &sync.RWMutex{},
			},
			args: args{
				in0:         context.TODO(),
				delegations: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				delegations: tt.fields.delegations,
				nextID:      tt.fields.nextID,
				mu:          tt.fields.mu,
			}
			if err := m.SaveDelegations(tt.args.in0, tt.args.delegations); (err != nil) != tt.wantErr {
				t.Errorf("SaveDelegationsWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_New(t *testing.T) {
	tests := []struct {
		name string
		want database.Adapter
	}{
		{
			name: "Nominal case",
			want: &Memory{
				delegations: make([]*model.Delegation, 0),
				nextID:      1,
			},
		},
		{
			name: "Error case - nil adapter",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
