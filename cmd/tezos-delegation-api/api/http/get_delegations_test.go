package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/tezos-delegation-service/internal/model"
	"github.com/tezos-delegation-service/internal/usecase"
)

func Test_GetDelegationHandler_GetDelegations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		getDelegationsFunc usecase.GetDelegationsFunc
		setupContext       func(*gin.Context)
		expectedStatus     int
		expectedHeader     map[string]string
		expectedError      string
		expectedCalled     bool
	}{
		{
			name: "nominal case",
			getDelegationsFunc: func(ctx context.Context, page, limit, year string, maxID int64) (*model.DelegationResponse, error) {
				return &model.DelegationResponse{
					Data: []model.Delegation{{ID: 1}},
					Pagination: model.PaginationInfo{
						CurrentPage: 1,
						PerPage:     50,
						HasNextPage: false,
					},
					MaxDelegationID: 1,
				}, nil
			},
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/?page=1&limit=50", nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: map[string]string{
				"X-Page-Current":  "1",
				"X-Page-Per-Page": "50",
				"Cache-Control":   "public, max-age=300",
			},
			expectedCalled: true,
		},
		{
			name: "error - invalid page",
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/?page=invalid", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid page number",
			expectedCalled: false,
		},
		{
			name: "error - limit out of bounds",
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/?limit=200", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "limit must be between 1 and 100, got 200",
			expectedCalled: false,
		},
		{
			name: "error - internal service",
			getDelegationsFunc: func(ctx context.Context, page, limit, year string, maxID int64) (*model.DelegationResponse, error) {
				return nil, errors.New("internal error")
			},
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/?page=1&limit=50", nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal error",
			expectedCalled: true,
		},
		/* {
			name: "cache hit - not modified",
			getDelegationsFunc: func(ctx context.Context, page, limit, year string, maxID int64) (*model.DelegationResponse, error) {
				return &model.DelegationResponse{
					Data: []model.Delegation{
						{ID: 1},
						{ID: 2},
					},
				}, nil
			},
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/?page=1&limit=50", nil)
				c.Request.Header.Set("If-None-Match", `"0b2c321dd2ca3add2f21a02b646398f3daed6b82f43fd5ba663631c1e9e3e903"`)
			},
			expectedStatus: http.StatusNotModified,
			expectedCalled: true,
		}, */
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			h := &GetDelegationHandler{
				getDelegationsFunc: tt.getDelegationsFunc,
			}
			h.GetDelegations(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}

			if tt.expectedHeader != nil {
				for key, value := range tt.expectedHeader {
					assert.Equal(t, value, w.Header().Get(key))
				}
			}
		})
	}
}

func Test_GetDelegationHandler_extractMaxDelegationID(t *testing.T) {
	tests := []struct {
		name string
		c    *gin.Context
		want int64
	}{
		{
			name: "Nominal case - valid header",
			c: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/", nil)
				c.Request.Header.Set("X-Max-Delegation-ID", "42")
				return c
			}(),
			want: 42,
		},
		{
			name: "Header absent",
			c: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/", nil)
				return c
			}(),
			want: 0,
		},
		{
			name: "Header with invalid value",
			c: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/", nil)
				c.Request.Header.Set("X-Max-Delegation-ID", "invalid")
				return c
			}(),
			want: 0,
		},
		{
			name: "Header with negative value",
			c: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/", nil)
				c.Request.Header.Set("X-Max-Delegation-ID", "-10")
				return c
			}(),
			want: 0,
		},
		{
			name: "Header with zero value",
			c: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/", nil)
				c.Request.Header.Set("X-Max-Delegation-ID", "0")
				return c
			}(),
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, (&GetDelegationHandler{}).extractMaxDelegationID(tt.c), "extractMaxDelegationID(%v)", tt.c)
		})
	}
}

func Test_GetDelegationHandler_setCacheHeaders(t *testing.T) {
	type args struct {
		c    *gin.Context
		year string
	}
	tests := []struct {
		name           string
		args           args
		expectedHeader string
	}{
		{
			name: "With specified year - 1h cache",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request, _ = http.NewRequest("GET", "/", nil)
					return c
				}(),
				year: "2023",
			},
			expectedHeader: "public, max-age=3600",
		},
		{
			name: "No year specified - 5m cache",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request, _ = http.NewRequest("GET", "/", nil)
					return c
				}(),
			},
			expectedHeader: "public, max-age=300",
		},
		{
			name: "With empty year - 5m cache",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request, _ = http.NewRequest("GET", "/", nil)
					return c
				}(),
			},
			expectedHeader: "public, max-age=300",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			(&GetDelegationHandler{}).setCacheHeaders(tt.args.c, tt.args.year)

			assert.Equal(t, tt.expectedHeader, tt.args.c.Writer.Header().Get("Cache-Control"))
		})
	}
}

func Test_GetDelegationHandler_setETagHeader(t *testing.T) {
	tests := []struct {
		name           string
		response       *model.DelegationResponse
		expectedPrefix string
		expectedLen    int
	}{
		{
			name: "Nominal case - complete response object",
			response: &model.DelegationResponse{
				Data: []model.Delegation{
					{ID: 1, Amount: 100},
					{ID: 2, Amount: 200},
				},
				Pagination: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     10,
				},
				MaxDelegationID: 100,
			},
			expectedPrefix: "\"",
			expectedLen:    66,
		},
		{
			name:           "Empty response object",
			response:       &model.DelegationResponse{},
			expectedPrefix: "\"",
			expectedLen:    66,
		},
		{
			name:           "Response object nil",
			response:       nil,
			expectedPrefix: "\"",
			expectedLen:    66,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/", nil)

			h := &GetDelegationHandler{}
			h.setETagHeader(c, tt.response)

			etag := w.Header().Get("ETag")
			assert.NotEmpty(t, etag)
			assert.True(t, len(etag) == tt.expectedLen, "ETag should have length %d but has %d (%s)", tt.expectedLen, len(etag), etag)
			assert.True(t, etag[0:1] == tt.expectedPrefix, "ETag should start with %s", tt.expectedPrefix)
			assert.True(t, etag[len(etag)-1:] == "\"", "ETag should end with '\"'")

			if tt.response != nil {
				w2 := httptest.NewRecorder()
				c2, _ := gin.CreateTestContext(w2)
				c2.Request, _ = http.NewRequest("GET", "/", nil)
				h.setETagHeader(c2, tt.response)
				assert.Equal(t, etag, w2.Header().Get("ETag"), "The same object should generate the same ETag")
			}
		})
	}
}

func Test_GetDelegationHandler_setMaxDelegationIDHeader(t *testing.T) {
	type args struct {
		c               *gin.Context
		maxDelegationID int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Nominal case - valid maxDelegationID",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request, _ = http.NewRequest("GET", "/", nil)
					return c
				}(),
				maxDelegationID: 42,
			},
		},
		{
			name: "Error case - maxDelegationID is zero",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request, _ = http.NewRequest("GET", "/", nil)
					return c
				}(),
				maxDelegationID: 0,
			},
		},
		{
			name: "Error case - maxDelegationID is negative",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request, _ = http.NewRequest("GET", "/", nil)
					return c
				}(),
				maxDelegationID: -10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			(&GetDelegationHandler{}).setMaxDelegationIDHeader(tt.args.c, tt.args.maxDelegationID)
		})
	}
}

func Test_GetDelegationHandler_setPaginationHeaders(t *testing.T) {
	type args struct {
		c     *gin.Context
		pInfo model.PaginationInfo
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Nominal case",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					return c
				}(),
				pInfo: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     50,
					HasPrevPage: true,
					PrevPage:    0,
					HasNextPage: true,
					NextPage:    2,
				},
			},
		},
		{
			name: "Error case - no previous or next page",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					return c
				}(),
				pInfo: model.PaginationInfo{
					CurrentPage: 1,
					PerPage:     50,
					HasPrevPage: false,
					HasNextPage: false,
				},
			},
		},
		{
			name: "Error case - invalid current page",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					return c
				}(),
				pInfo: model.PaginationInfo{
					CurrentPage: -1,
					PerPage:     50,
					HasPrevPage: false,
					HasNextPage: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			(&GetDelegationHandler{}).setPaginationHeaders(tt.args.c, tt.args.pInfo)
		})
	}
}

func Test_GetDelegationHandler_setRequestIDHeader(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Nominal case - request ID already set",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request, _ = http.NewRequest("GET", "/", nil)
					c.Request.Header.Set("X-Request-ID", "existing-request-id")
					return c
				}(),
			},
		},
		{
			name: "Nominal case - request ID not set",
			args: args{
				c: func() *gin.Context {
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					c.Request, _ = http.NewRequest("GET", "/", nil)
					return c
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			(&GetDelegationHandler{}).setRequestIDHeader(tt.args.c)
		})
	}
}

func Test_GetDelegationHandler_validateRequestParams(t *testing.T) {
	tests := []struct {
		name    string
		c       *gin.Context
		want    int
		want1   int
		want2   string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "nominal case",
			c: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/?page=1&limit=50&year=2023", nil)
				return c
			}(),
			want:    1,
			want1:   50,
			want2:   "2023",
			wantErr: assert.NoError,
		},
		{
			name: "error - invalid page",
			c: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/?page=invalid", nil)
				return c
			}(),
			want:    0,
			want1:   0,
			want2:   "",
			wantErr: assert.Error,
		},
		{
			name: "error - limit out of bounds",
			c: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/?limit=200", nil)
				return c
			}(),
			want:    0,
			want1:   0,
			want2:   "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := (&GetDelegationHandler{}).validateRequestParams(tt.c)
			if !tt.wantErr(t, err, fmt.Sprintf("validateRequestParams(%v)", tt.c)) {
				return
			}
			assert.Equalf(t, tt.want, got, "validateRequestParams(%v)", tt.c)
			assert.Equalf(t, tt.want1, got1, "validateRequestParams(%v)", tt.c)
			assert.Equalf(t, tt.want2, got2, "validateRequestParams(%v)", tt.c)
		})
	}
}

func Test_NewGetDelegationHandler(t *testing.T) {
	tests := []struct {
		name               string
		getDelegationsFunc usecase.GetDelegationsFunc
		want               *GetDelegationHandler
	}{
		{
			name: "nominal case",
			getDelegationsFunc: func(ctx context.Context, page, limit, year string, maxID int64) (*model.DelegationResponse, error) {
				return &model.DelegationResponse{}, nil
			},
			want: nil,
		},
		{
			name:               "error case - nil function",
			getDelegationsFunc: nil,
			want: &GetDelegationHandler{
				getDelegationsFunc: nil,
			},
		},
		{
			name: "error case - function returns error",
			getDelegationsFunc: func(ctx context.Context, page, limit, year string, maxID int64) (*model.DelegationResponse, error) {
				return nil, errors.New("internal error")
			},
			want: &GetDelegationHandler{
				getDelegationsFunc: func(ctx context.Context, page, limit, year string, maxID int64) (*model.DelegationResponse, error) {
					return nil, errors.New("internal error")
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewGetDelegationHandler(uint16(50), tt.getDelegationsFunc)
			assert.NotNil(t, handler)

			if tt.name == "nominal case" {
				resp, err := handler.getDelegationsFunc(context.Background(), "1", "10", "2023", 0)
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else if tt.name == "error case - nil function" {
				assert.Nil(t, handler.getDelegationsFunc)
			} else if tt.name == "error case - function returns error" {
				_, err := handler.getDelegationsFunc(context.Background(), "1", "10", "2023", 0)
				assert.Error(t, err)
				assert.Equal(t, "internal error", err.Error())
			}
		})
	}
}
