package usecase

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/adapter/database"
	dbmock "github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	"github.com/tezos-delegation-service/internal/model"
)

func Test_NewGetDelegationsFunc(t *testing.T) {
	providedDBAdapter := dbmock.New()
	providedMetricsClient := metricsnoop.New()

	type args struct {
		adapter       database.Adapter
		defaultLimit  uint16
		metricsClient metrics.Adapter
	}
	tests := []struct {
		name  string
		args  args
		valid bool
	}{
		{
			name: "Nominal case",
			args: args{
				adapter:       providedDBAdapter,
				defaultLimit:  50,
				metricsClient: providedMetricsClient,
			},
			valid: true,
		},
		{
			name: "Nil adapter",
			args: args{
				adapter:       nil,
				defaultLimit:  50,
				metricsClient: providedMetricsClient,
			},
			valid: true,
		},
		{
			name: "Nil metrics client",
			args: args{
				adapter:       providedDBAdapter,
				defaultLimit:  50,
				metricsClient: nil,
			},
			valid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGetDelegationsFunc(tt.args.defaultLimit, tt.args.adapter, tt.args.metricsClient)
			if tt.valid && got == nil {
				t.Errorf("NewGetDelegationsFunc() returned nil, expected valid function")
			} else if !tt.valid && got != nil {
				t.Errorf("NewGetDelegationsFunc() = %v, want nil", got)
			}
		})
	}
}

func Test_getDelegations_GetDelegations(t *testing.T) {
	type args struct {
		ctx             context.Context
		pageStr         string
		limitStr        string
		yearStr         string
		maxDelegationID int64
	}
	tests := []struct {
		name      string
		dbAdapter database.Adapter
		args      args
		want      *model.DelegationResponse
		wantErr   bool
	}{
		{
			name: "Nominal",
			dbAdapter: func() database.Adapter {
				mockDB := dbmock.New()
				delegations := []model.Delegation{
					{
						ID:        1,
						Delegator: "tz1...",
						Delegate:  "tz2...",
						Amount:    100.0,
						Timestamp: func() int64 {
							t, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
							return t.Unix()
						}(),
						Level: 1000,
					},
				}
				mockDB.On("GetDelegations", mock.Anything, uint32(1), uint16(10), uint16(2025), uint64(0)).
					Return(delegations, nil)
				return mockDB
			}(),
			args: args{
				ctx:             context.Background(),
				pageStr:         "1",
				limitStr:        "10",
				yearStr:         "2025",
				maxDelegationID: 0,
			},
			want: &model.DelegationResponse{
				Data: []model.Delegation{
					{
						ID:        1,
						Delegator: "tz1...",
						Delegate:  "tz2...",
						Amount:    100000000.0, // Converti en mutez
						Timestamp: func() int64 {
							t, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
							return t.Unix()
						}(),
						TimestampTime: "2025-01-01T00:00:00Z",
						Level:         1000,
					},
				},
				Pagination: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     10,
					HasPrevPage: false,
					HasNextPage: false,
				},
				MaxDelegationID: 1,
			},
			wantErr: false,
		},
		{
			name: "With maxDelegationID",
			dbAdapter: func() database.Adapter {
				mockDB := dbmock.New()
				delegations := []model.Delegation{
					{
						ID:        50,
						Delegator: "tz1...",
						Delegate:  "tz2...",
						Amount:    100.0,
						Timestamp: func() int64 {
							t, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
							return t.Unix()
						}(),
						Level: 1000,
					},
				}
				mockDB.On("GetDelegations", mock.Anything, uint32(2), uint16(10), uint16(2025), uint64(100)).
					Return(delegations, 60, nil)
				return mockDB
			}(),
			args: args{
				ctx:             context.Background(),
				pageStr:         "2",
				limitStr:        "10",
				yearStr:         "2025",
				maxDelegationID: 100,
			},
			want: &model.DelegationResponse{
				Data: []model.Delegation{
					{
						ID:        50,
						Delegator: "tz1...",
						Delegate:  "tz2...",
						Amount:    100000000.0,
						Timestamp: func() int64 {
							t, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
							return t.Unix()
						}(),
						TimestampTime: "2025-01-01T00:00:00Z",
						Level:         1000,
					},
				},
				Pagination: model.PaginationInfo{
					CurrentPage: 2,
					PerPage:     10,
					HasPrevPage: true,
					HasNextPage: true,
					PrevPage:    1,
					NextPage:    3,
				},
				MaxDelegationID: 50,
			},
			wantErr: false,
		},
		{
			name: "Invalid pageStr",
			dbAdapter: func() database.Adapter {
				mockDB := dbmock.New()
				mockDB.On("GetDelegations", mock.Anything, uint32(1), uint16(10), uint16(0), uint64(0)).
					Return([]model.Delegation{}, nil)
				return mockDB
			}(),
			args: args{
				ctx:             context.Background(),
				pageStr:         "invalid",
				limitStr:        "10",
				yearStr:         "",
				maxDelegationID: 0,
			},
			want: &model.DelegationResponse{
				Data: []model.Delegation{},
				Pagination: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     10,
					HasPrevPage: false,
					HasNextPage: false,
				},
				MaxDelegationID: 0,
			},
			wantErr: true,
		},
		{
			name: "Invalid limitStr",
			dbAdapter: func() database.Adapter {
				mockDB := dbmock.New()
				mockDB.On("GetDelegations", mock.Anything, uint32(1), uint16(50), uint16(0), uint64(0)).
					Return([]model.Delegation{}, nil)
				return mockDB
			}(),
			args: args{
				ctx:             context.Background(),
				pageStr:         "1",
				limitStr:        "invalid",
				yearStr:         "",
				maxDelegationID: 0,
			},
			want: &model.DelegationResponse{
				Data: []model.Delegation{},
				Pagination: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     50,
					HasPrevPage: false,
					HasNextPage: false,
				},
				MaxDelegationID: 0,
			},
			wantErr: true,
		},
		{
			name: "Database error",
			dbAdapter: func() database.Adapter {
				mockDB := dbmock.New()
				mockDB.On("GetDelegations", mock.Anything, uint32(1), uint16(10), uint16(0), uint64(0)).
					Return([]model.Delegation{}, fmt.Errorf("database error"))
				return mockDB
			}(),
			args: args{
				ctx:             context.Background(),
				pageStr:         "1",
				limitStr:        "10",
				yearStr:         "",
				maxDelegationID: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &getDelegations{
				dbAdapter: tt.dbAdapter,
			}
			got, err := uc.GetDelegations(tt.args.ctx, tt.args.pageStr, tt.args.limitStr, tt.args.yearStr, tt.args.maxDelegationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDelegations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDelegations() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDelegations_parseLimit(t *testing.T) {
	tests := []struct {
		name     string
		limitStr string
		want     uint16
		wantErr  bool
	}{
		{
			name:     "Nominal case",
			limitStr: "10",
			want:     10,
			wantErr:  false,
		},
		{
			name:     "Empty limitStr",
			limitStr: "",
			want:     50,
			wantErr:  false,
		},
		{
			name:     "Invalid limitStr",
			limitStr: "invalid",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "Negative limitStr",
			limitStr: "-10",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "Zero limitStr",
			limitStr: "0",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "Exceeds maximum value",
			limitStr: "70000",
			want:     0,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&getDelegations{
				defaultLimit: 50,
			}).parseLimit(tt.limitStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseLimit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("parseLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDelegations_parsePage(t *testing.T) {
	tests := []struct {
		name    string
		pageStr string
		want    uint32
		wantErr bool
	}{
		{
			name:    "Nominal case",
			pageStr: "2",
			want:    2,
			wantErr: false,
		},
		{
			name:    "Empty pageStr",
			pageStr: "",
			want:    1,
			wantErr: false,
		},
		{
			name:    "Invalid pageStr",
			pageStr: "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Negative pageStr",
			pageStr: "-1",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Zero pageStr",
			pageStr: "0",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Exceeds maximum value",
			pageStr: "4294967296", // > max uint32
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&getDelegations{}).parsePage(tt.pageStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("parsePage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDelegations_parseYear(t *testing.T) {
	tests := []struct {
		name     string
		yearStr  string
		want     uint16
		wantErr  bool
		errorMsg string
	}{
		{
			name:    "Nominal case",
			yearStr: "2025",
			want:    2025,
			wantErr: false,
		},
		{
			name:    "Empty yearStr",
			yearStr: "",
			want:    0,
			wantErr: false,
		},
		{
			name:     "Invalid yearStr",
			yearStr:  "invalid",
			want:     0,
			wantErr:  true,
			errorMsg: "strconv.Atoi: parsing \"invalid\": invalid syntax",
		},
		{
			name:     "Negative yearStr",
			yearStr:  "-2025",
			want:     0,
			wantErr:  true,
			errorMsg: "year must be a positive number",
		},
		{
			name:     "Zero yearStr",
			yearStr:  "0",
			want:     0,
			wantErr:  true,
			errorMsg: "year must be a positive number",
		},
		{
			name:     "Future year",
			yearStr:  fmt.Sprintf("%d", time.Now().Year()+1),
			want:     0,
			wantErr:  true,
			errorMsg: "year cannot exceed the current year",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&getDelegations{}).parseYear(tt.yearStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseYear() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("parseYear() error = %v, wantErrMsg %v", err, tt.errorMsg)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("parseYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDelegations_withMonitorer(t *testing.T) {
	type fields struct {
		dbAdapter database.Adapter
	}
	type args struct {
		getDelegations GetDelegationsFunc
		metricsClient  metrics.Adapter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   GetDelegationsFunc
	}{
		{
			name: "Nominal case",
			fields: fields{
				dbAdapter: dbmock.New(),
			},
			args: args{
				getDelegations: func(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.DelegationResponse, error) {
					return &model.DelegationResponse{}, nil
				},
				metricsClient: metricsnoop.New(),
			},
			want: func(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.DelegationResponse, error) {
				return &model.DelegationResponse{}, nil
			},
		},
		{
			name: "Error case - nil metrics client",
			fields: fields{
				dbAdapter: dbmock.New(),
			},
			args: args{
				getDelegations: func(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.DelegationResponse, error) {
					return nil, nil
				},
				metricsClient: nil,
			},
			want: func(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.DelegationResponse, error) {
				return nil, nil
			},
		},
		{
			name: "Error case - getDelegations returns error",
			fields: fields{
				dbAdapter: dbmock.New(),
			},
			args: args{
				getDelegations: func(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.DelegationResponse, error) {
					return nil, fmt.Errorf("error")
				},
				metricsClient: metricsnoop.New(),
			},
			want: func(ctx context.Context, pageStr, limitStr, yearStr string, maxDelegationID int64) (*model.DelegationResponse, error) {
				return nil, fmt.Errorf("error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &getDelegations{
				dbAdapter: tt.fields.dbAdapter,
			}
			got := uc.withMonitorer(tt.args.getDelegations, tt.args.metricsClient)

			if tt.want == nil && got != nil {
				t.Errorf("withMonitorer() = %v, want nil", got)
				return
			}
			if tt.want != nil && got == nil {
				t.Errorf("withMonitorer() returned nil, expected a function")
				return
			}

			if tt.want != nil && got != nil {
				ctx := context.Background()
				gotResp, gotErr := got(ctx, "1", "10", "2025", 0)
				wantResp, wantErr := tt.want(ctx, "1", "10", "2025", 0)

				if (gotErr == nil) != (wantErr == nil) {
					t.Errorf("withMonitorer() error = %v, want error = %v", gotErr, wantErr)
					return
				}

				if !reflect.DeepEqual(gotResp, wantResp) {
					t.Errorf("withMonitorer() response = %v, want %v", gotResp, wantResp)
				}
			}
		})
	}
}
