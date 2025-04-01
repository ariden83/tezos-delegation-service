package mock

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

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

func Test_Mock_FetchDelegations(t *testing.T) {
	type args struct {
		ctx    context.Context
		limit  int
		offset int
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		want    model.TzktDelegationResponse
		wantErr bool
	}{
		{
			name: "Nominal case",
			mock: func() *Mock {
				m := New()
				m.On("FetchDelegations", mock.Anything, 10, 0).
					Return(stubTZKTDelegationResponse, nil)
				return m
			}(),
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
			},
			want:    stubTZKTDelegationResponse,
			wantErr: false,
		},
		{
			name: "Error case - API error",
			mock: func() *Mock {
				m := New()
				m.On("FetchDelegations", mock.Anything, 10, 0).
					Return(model.TzktDelegationResponse{}, errors.New("API error"))
				return m
			}(),
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
			},
			want:    model.TzktDelegationResponse{},
			wantErr: true,
		},
		{
			name: "Error case - Invalid response",
			mock: func() *Mock {
				m := New()
				m.On("FetchDelegations", mock.Anything, 10, 0).
					Return(model.TzktDelegationResponse{}, errors.New("invalid response"))
				return m
			}(),
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
			},
			want:    model.TzktDelegationResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			got, err := m.FetchDelegations(tt.args.ctx, tt.args.limit, tt.args.offset)
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

func Test_Mock_FetchDelegationsFromLevel(t *testing.T) {
	type args struct {
		ctx   context.Context
		level int64
	}
	tests := []struct {
		name    string
		mock    *Mock
		args    args
		want    model.TzktDelegationResponse
		wantErr bool
	}{
		{
			name: "Nominal case",
			mock: func() *Mock {
				m := New()
				m.On("FetchDelegationsFromLevel", mock.Anything, int64(100)).
					Return(stubTZKTDelegationResponse, nil)
				return m
			}(),
			args: args{
				ctx:   context.TODO(),
				level: 100,
			},
			want:    stubTZKTDelegationResponse,
			wantErr: false,
		},
		{
			name: "Error case - API error",
			mock: func() *Mock {
				m := New()
				m.On("FetchDelegationsFromLevel", mock.Anything, int64(100)).
					Return(model.TzktDelegationResponse{}, errors.New("API error"))
				return m
			}(),
			args: args{
				ctx:   context.TODO(),
				level: 100,
			},
			want:    model.TzktDelegationResponse{},
			wantErr: true,
		},
		{
			name: "Error case - Invalid response",
			mock: func() *Mock {
				m := New()
				m.On("FetchDelegationsFromLevel", mock.Anything, int64(100)).
					Return(model.TzktDelegationResponse{}, errors.New("invalid response"))
				return m
			}(),
			args: args{
				ctx:   context.TODO(),
				level: 100,
			},
			want:    model.TzktDelegationResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.mock
			got, err := m.FetchDelegationsFromLevel(tt.args.ctx, tt.args.level)
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

func Test_New(t *testing.T) {
	providedMock := &Mock{}
	t.Run("nominal", func(t *testing.T) {
		if got := New(); !reflect.DeepEqual(got, providedMock) {
			t.Errorf("New() = %v, want %v", got, providedMock)
		}
	})
}
