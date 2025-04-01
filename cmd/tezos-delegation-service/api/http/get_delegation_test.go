package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/model"
)

// Mock service
type MockDelegationService struct {
	mock.Mock
}

func (m *MockDelegationService) GetDelegations(ctx context.Context, page string, limit string, year string) (*model.DelegationResponse, error) {
	args := m.Called(ctx, page, limit, year)
	return args.Get(0).(*model.DelegationResponse), args.Error(1)
}

func TestGetDelegations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock service
	mockService := new(MockDelegationService)

	// Create the handler
	handler := NewDelegationHandler(mockService)

	// Create a router for testing
	router := gin.Default()
	router.GET("/xtz/delegations", handler.GetDelegations)

	// Set up mock data
	testTime := time.Now().Unix()
	testResponse := &model.DelegationResponse{
		Data: []model.Delegation{
			{
				Delegator: "tz1a1SAaXRt9yoGMx29rh9FsBF4UzmvojdTL",
				Timestamp: testTime,
				Amount:    100.5,
				Level:     2338084,
			},
		},
		Pagination: model.PaginationInfo{
			CurrentPage: 1,
			PerPage:     50,
			TotalItems:  1,
			TotalPages:  1,
			HasPrevPage: false,
			HasNextPage: false,
		},
	}

	// Test without query parameters
	mockService.On("GetDelegations", mock.Anything, "1", "50", "").Return(testResponse, nil)

	// Create a test request
	req, _ := http.NewRequest("GET", "/xtz/delegations", nil)
	resp := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(resp, req)

	// Check the response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Check cache headers for general requests
	assert.Equal(t, "public, max-age=300", resp.Header().Get("Cache-Control"))
	assert.NotEmpty(t, resp.Header().Get("ETag"))

	var response model.DelegationResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, "tz1a1SAaXRt9yoGMx29rh9FsBF4UzmvojdTL", response.Data[0].Delegator)

	// Test with year parameter
	mockService.On("GetDelegations", mock.Anything, "2", "25", "2022").Return(testResponse, nil)

	// Create a test request with parameters
	req, _ = http.NewRequest("GET", "/xtz/delegations?page=2&limit=25&year=2022", nil)
	resp = httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(resp, req)

	// Check the response
	assert.Equal(t, http.StatusOK, resp.Code)

	// Check cache headers for year-specific requests
	assert.Equal(t, "public, max-age=3600", resp.Header().Get("Cache-Control"))
	assert.NotEmpty(t, resp.Header().Get("ETag"))

	err = json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Data, 1)

	// Test conditional request with If-None-Match header
	// First, get an ETag from a normal request
	req, _ = http.NewRequest("GET", "/xtz/delegations", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	etag := resp.Header().Get("ETag")

	// Now make a conditional request with the ETag
	req, _ = http.NewRequest("GET", "/xtz/delegations", nil)
	req.Header.Set("If-None-Match", etag)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Should get a 304 Not Modified response
	assert.Equal(t, http.StatusNotModified, resp.Code)
	assert.Empty(t, resp.Body.String()) // Body should be empty for 304

	mockService.AssertExpectations(t)
}
