package api

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/tzktapi"
	"github.com/tezos-delegation-service/internal/model"
)

// httpClientMock returns a mock HTTP client that uses the provided function to return responses.
func httpClientMock(fn func(*http.Request) *http.Response) *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(fn),
	}
}

type roundTripperFunc func(*http.Request) *http.Response

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func Test_Adapter_FetchDelegations(t *testing.T) {
	type fields struct {
		apiURL string
		client *http.Client
		db     database.Adapter
		logger *logrus.Entry
	}
	type args struct {
		ctx    context.Context
		limit  uint16
		offset int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TzktDelegationResponse
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`[{"level": 1000, "id": 123}]`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:    context.Background(),
				limit:  10,
				offset: 0,
			},
			want: model.TzktDelegationResponse{{
				Level: 1000,
				ID:    123,
			}},
			wantErr: false,
		},
		{
			name: "Error case - invalid URL",
			fields: fields{
				apiURL: "http://invalid-url",
				client: &http.Client{Timeout: 30 * time.Second},
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:    context.Background(),
				limit:  10,
				offset: 0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error case - context cancelled",
			fields: fields{
				apiURL: "http://example.com",
				client: &http.Client{Timeout: 30 * time.Second},
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				limit:  10,
				offset: 0,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				apiURL: tt.fields.apiURL,
				client: tt.fields.client,
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			got, err := a.FetchDelegations(tt.args.ctx, tt.args.limit, tt.args.offset)
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

func Test_Adapter_FetchDelegationsFromLevel(t *testing.T) {
	type fields struct {
		apiURL string
		client *http.Client
		db     database.Adapter
		logger *logrus.Entry
	}
	type args struct {
		ctx   context.Context
		level uint64
		limit uint8
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TzktDelegationResponse
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`[{"level": 1001}]`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:   context.Background(),
				level: 1000,
				limit: 100,
			},
			want: model.TzktDelegationResponse{{
				Level: 1001,
			}},
			wantErr: false,
		},
		{
			name: "Error case - invalid URL",
			fields: fields{
				apiURL: "http://invalid-url",
				client: &http.Client{Timeout: 30 * time.Second},
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:   context.Background(),
				level: 1000,
				limit: 100,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error case - context cancelled",
			fields: fields{
				apiURL: "http://example.com",
				client: &http.Client{Timeout: 30 * time.Second},
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				level: 1000,
				limit: 100,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				apiURL: tt.fields.apiURL,
				client: tt.fields.client,
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			got, err := a.FetchDelegationsFromLevel(tt.args.ctx, tt.args.level, tt.args.limit)
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

func Test_Adapter_GetCurrentCycle(t *testing.T) {
	type fields struct {
		apiURL string
		client *http.Client
		db     database.Adapter
		logger *logrus.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		ctx     context.Context
		want    int
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"cycle": 42}`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			ctx:     context.Background(),
			want:    42,
			wantErr: false,
		},
		{
			name: "Error case - API error",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(strings.NewReader(`{"error": "server error"}`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			ctx:     context.Background(),
			want:    0,
			wantErr: true,
		},
		{
			name: "Error case - invalid JSON",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`invalid json`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			ctx:     context.Background(),
			want:    0,
			wantErr: true,
		},
		{
			name: "Error case - context cancelled",
			fields: fields{
				apiURL: "http://example.com",
				client: &http.Client{Timeout: 30 * time.Second},
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				apiURL: tt.fields.apiURL,
				client: tt.fields.client,
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			got, err := a.GetCurrentCycle(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentCycle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetCurrentCycle() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Adapter_FetchRewardsForCycle(t *testing.T) {
	type fields struct {
		apiURL string
		client *http.Client
		db     database.Adapter
		logger *logrus.Entry
	}
	type args struct {
		ctx       context.Context
		delegator model.WalletAddress
		baker     model.WalletAddress
		cycle     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.Reward
		wantErr bool
	}{
		{
			name: "Nominal case",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(strings.NewReader(`{
							"rewardsShare": 5.5,
							"baker": {"address": "tz1baker1"},
							"cycle": 10,
							"timestamp": 1672531199
						}`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:       context.Background(),
				delegator: "tz1delegator1",
				baker:     "tz1baker1",
				cycle:     10,
			},
			want: []model.Reward{
				{
					RecipientAddress: "tz1delegator1",
					SourceAddress:    "tz1baker1",
					Cycle:            10,
					Amount:           5.5,
					Timestamp:        1672531199,
					TimestampTime:    time.Unix(1672531199, 0).Format(time.RFC3339),
				},
			},
			wantErr: false,
		},
		{
			name: "Case with zero reward",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(strings.NewReader(`{
							"rewardsShare": 0,
							"baker": {"address": "tz1baker1"},
							"cycle": 10,
							"timestamp": 1672531199
						}`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:       context.Background(),
				delegator: "tz1delegator1",
				baker:     "tz1baker1",
				cycle:     10,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Error case - API error",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(strings.NewReader(`{"error": "server error"}`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:       context.Background(),
				delegator: "tz1delegator1",
				baker:     "tz1baker1",
				cycle:     10,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error case - invalid JSON",
			fields: fields{
				apiURL: "http://example.com",
				client: httpClientMock(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`invalid json`)),
					}
				}),
				db:     nil,
				logger: logrus.NewEntry(logrus.New()),
			},
			args: args{
				ctx:       context.Background(),
				delegator: "tz1delegator1",
				baker:     "tz1baker1",
				cycle:     10,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				apiURL: tt.fields.apiURL,
				client: tt.fields.client,
				db:     tt.fields.db,
				logger: tt.fields.logger,
			}
			got, err := a.FetchRewardsForCycle(tt.args.ctx, tt.args.delegator, tt.args.baker, tt.args.cycle)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchRewardsForCycle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchRewardsForCycle() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_New(t *testing.T) {
	type args struct {
		cfg    Config
		logger *logrus.Entry
	}
	tests := []struct {
		name    string
		args    args
		want    tzktapi.Adapter
		wantErr bool
	}{
		{
			name: "Nominal case",
			args: args{
				cfg: Config{
					URL:     "http://example.com",
					Timeout: 30 * time.Second,
				},
				logger: logrus.NewEntry(logrus.New()),
			},
			want: &Adapter{
				apiURL: "http://example.com",
				client: &http.Client{Timeout: 30 * time.Second},
				logger: logrus.NewEntry(logrus.New()),
			},
			wantErr: false,
		},
		{
			name: "Error case - missing URL",
			args: args{
				cfg: Config{
					URL:     "",
					Timeout: 30 * time.Second,
				},
				logger: logrus.NewEntry(logrus.New()),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Zero timeout case",
			args: args{
				cfg: Config{
					URL:     "http://example.com",
					Timeout: 0,
				},
				logger: logrus.NewEntry(logrus.New()),
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				adapter, ok := got.(*Adapter)
				if !ok {
					t.Error("Expected *Adapter type")
					return
				}
				if adapter.apiURL != tt.args.cfg.URL {
					t.Errorf("New() got apiURL = %v, want %v", adapter.apiURL, tt.args.cfg.URL)
				}
				if adapter.client.Timeout != 30*time.Second {
					t.Errorf("New() got timeout = %v, want %v", adapter.client.Timeout, 30*time.Second)
				}
			}
		})
	}
}
