package usecase

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/adapter/metrics"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	tzktmock "github.com/tezos-delegation-service/internal/adapter/tzktapi/impl/mock"
	"github.com/tezos-delegation-service/internal/model"
)

func Test_NewGetDelegationsFunc(t *testing.T) {
	providedTZKTAPI := tzktmock.New()
	providedMetricsClient := metricsnoop.New()

	type args struct {
		adapter       tzktapi.Adapter
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
				adapter:       providedTZKTAPI,
				metricsClient: providedMetricsClient,
			},
			valid: true,
		},
		{
			name: "Nil adapter",
			args: args{
				adapter:       nil,
				metricsClient: providedMetricsClient,
			},
			valid: true,
		},
		{
			name: "Nil metrics client",
			args: args{
				adapter:       providedTZKTAPI,
				metricsClient: nil,
			},
			valid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGetDelegationsFunc(tt.args.adapter, tt.args.metricsClient)
			if tt.valid && got == nil {
				t.Errorf("NewGetDelegationsFunc() returned nil, expected valid function")
			} else if !tt.valid && got != nil {
				t.Errorf("NewGetDelegationsFunc() = %v, want nil", got)
			}
		})
	}
}

func Test_getDelegations_GetDelegations(t *testing.T) {
	type fields struct {
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		ctx      context.Context
		pageStr  string
		limitStr string
		yearStr  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.DelegationResponse
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				tzktApiAdapter: func() tzktapi.Adapter {
					providedTZKTAPI := tzktmock.New()
					providedTZKTAPI.On("FetchDelegations", mock.Anything, 10, 0).
						Return(model.TzktDelegationResponse{
							{
								Sender:   model.TzktAddress{Address: "tz1..."},
								Delegate: model.TzktDelegate{Address: "tz2..."},
								Amount:   100000000,
								Status:   "applied",
								Timestamp: func() time.Time {
									t, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
									return t
								}(),
								Level: 1000,
							},
						}, nil)
					return providedTZKTAPI
				}(),
			},
			args: args{
				ctx:      context.Background(),
				pageStr:  "1",
				limitStr: "10",
				yearStr:  "2023",
			},
			want: &model.DelegationResponse{
				Data: []model.Delegation{
					{
						Delegator: "tz1...",
						Delegate:  "tz2...",
						Amount:    100.0,
						Timestamp: func() int64 {
							t, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
							return t.Unix()
						}(),
						Level: 1000,
					},
				},
				Pagination: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     10,
					HasPrevPage: false,
					HasNextPage: false,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid pageStr",
			fields: fields{
				tzktApiAdapter: func() tzktapi.Adapter {
					providedTZKTAPI := tzktmock.New()
					providedTZKTAPI.On("FetchDelegations", mock.Anything, 10, 0).
						Return(model.TzktDelegationResponse{
							{Amount: 100, Status: "applied", Timestamp: time.Now()},
							{Amount: 120, Status: "applied", Timestamp: time.Now()},
							{Amount: 140, Status: "applied", Timestamp: time.Now()},
						}, nil)
					return providedTZKTAPI
				}(),
			},
			args: args{
				ctx:      context.Background(),
				pageStr:  "invalid",
				limitStr: "10",
				yearStr:  "2023",
			},
			want: &model.DelegationResponse{
				Data: []model.Delegation{},
				Pagination: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     10,
					HasPrevPage: false,
					HasNextPage: false,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid limitStr",
			fields: fields{
				tzktApiAdapter: func() tzktapi.Adapter {
					providedTZKTAPI := tzktmock.New()
					providedTZKTAPI.On("FetchDelegations", mock.Anything, 50, 0).
						Return(model.TzktDelegationResponse{
							{Amount: 100, Status: "applied", Timestamp: time.Now()},
							{Amount: 120, Status: "applied", Timestamp: time.Now()},
							{Amount: 140, Status: "applied", Timestamp: time.Now()},
						}, nil)
					return providedTZKTAPI
				}(),
			},
			args: args{
				ctx:      context.Background(),
				pageStr:  "1",
				limitStr: "invalid",
				yearStr:  "2023",
			},
			want: &model.DelegationResponse{
				Data: []model.Delegation{},
				Pagination: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     50,
					HasPrevPage: false,
					HasNextPage: false,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid yearStr",
			fields: fields{
				tzktApiAdapter: func() tzktapi.Adapter {
					providedTZKTAPI := tzktmock.New()
					providedTZKTAPI.On("FetchDelegations", mock.Anything, 10, 0).
						Return(model.TzktDelegationResponse{
							{
								Sender:    model.TzktAddress{Address: "tz1..."},
								Delegate:  model.TzktDelegate{Address: "tz2..."},
								Amount:    100000000,
								Status:    "applied",
								Timestamp: time.Now(),
								Level:     1000,
							},
						}, nil)
					return providedTZKTAPI
				}(),
			},
			args: args{
				ctx:      context.Background(),
				pageStr:  "1",
				limitStr: "10",
				yearStr:  "invalid",
			},
			want: &model.DelegationResponse{
				Data: []model.Delegation{
					{
						Delegator: "tz1...",
						Delegate:  "tz2...",
						Amount:    100.0,
						Timestamp: time.Now().Unix(),
						Level:     1000,
					},
				},
				Pagination: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     10,
					HasPrevPage: false,
					HasNextPage: false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &getDelegations{
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			got, err := uc.GetDelegations(tt.args.ctx, tt.args.pageStr, tt.args.limitStr, tt.args.yearStr)
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

func Test_getDelegations_parseLimit(t *testing.T) {
	type fields struct {
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		limitStr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Nominal case",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				limitStr: "10",
			},
			want: 10,
		},
		{
			name: "Empty limitStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				limitStr: "",
			},
			want: 50,
		},
		{
			name: "Invalid limitStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				limitStr: "invalid",
			},
			want: 50,
		},
		{
			name: "Negative limitStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				limitStr: "-10",
			},
			want: 50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &getDelegations{
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			if got := uc.parseLimit(tt.args.limitStr); got != tt.want {
				t.Errorf("parseLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDelegations_parsePage(t *testing.T) {
	type fields struct {
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		pageStr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Nominal case",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				pageStr: "2",
			},
			want: 2,
		},
		{
			name: "Empty pageStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				pageStr: "",
			},
			want: 1,
		},
		{
			name: "Invalid pageStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				pageStr: "invalid",
			},
			want: 1,
		},
		{
			name: "Negative pageStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				pageStr: "-1",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &getDelegations{
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			if got := uc.parsePage(tt.args.pageStr); got != tt.want {
				t.Errorf("parsePage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDelegations_parseYear(t *testing.T) {
	type fields struct {
		tzktApiAdapter tzktapi.Adapter
	}
	type args struct {
		yearStr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Nominal case",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				yearStr: "2023",
			},
			want: 2023,
		},
		{
			name: "Empty yearStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				yearStr: "",
			},
			want: 0,
		},
		{
			name: "Invalid yearStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				yearStr: "invalid",
			},
			want: 0,
		},
		{
			name: "Negative yearStr",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				yearStr: "-2023",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &getDelegations{
				tzktApiAdapter: tt.fields.tzktApiAdapter,
			}
			if got := uc.parseYear(tt.args.yearStr); got != tt.want {
				t.Errorf("parseYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDelegations_withMonitorer(t *testing.T) {
	type fields struct {
		tzktApiAdapter tzktapi.Adapter
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
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				getDelegations: func(ctx context.Context, pageStr, limitStr, yearStr string) (*model.DelegationResponse, error) {
					return &model.DelegationResponse{}, nil
				},
				metricsClient: metricsnoop.New(),
			},
			want: func(ctx context.Context, pageStr, limitStr, yearStr string) (*model.DelegationResponse, error) {
				return &model.DelegationResponse{}, nil
			},
		},
		{
			name: "Error case - nil metrics client",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				getDelegations: func(ctx context.Context, pageStr, limitStr, yearStr string) (*model.DelegationResponse, error) {
					return nil, nil
				},
				metricsClient: nil,
			},
			want: func(ctx context.Context, pageStr, limitStr, yearStr string) (*model.DelegationResponse, error) {
				return nil, nil
			},
		},
		{
			name: "Error case - getDelegations returns error",
			fields: fields{
				tzktApiAdapter: tzktmock.New(),
			},
			args: args{
				getDelegations: func(ctx context.Context, pageStr, limitStr, yearStr string) (*model.DelegationResponse, error) {
					return nil, fmt.Errorf("error")
				},
				metricsClient: metricsnoop.New(),
			},
			want: func(ctx context.Context, pageStr, limitStr, yearStr string) (*model.DelegationResponse, error) {
				return nil, fmt.Errorf("error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &getDelegations{
				tzktApiAdapter: tt.fields.tzktApiAdapter,
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
				gotResp, gotErr := got(ctx, "1", "10", "2023")
				wantResp, wantErr := tt.want(ctx, "1", "10", "2023")

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
