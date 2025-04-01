package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tezos-delegation-service/internal/model"
)

// Mock delegation service
type MockDelegationService struct {
	mock.Mock
}

func (m *MockDelegationService) GetDelegations(pageStr string, limitStr string, yearStr string) (*model.DelegationResponse, error) {
	args := m.Called(pageStr, limitStr, yearStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DelegationResponse), args.Error(1)
}

func TestNewServer(t *testing.T) {
	// Create a mock DB
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Create health service
	healthService := NewHealthService(sqlxDB)

	// Create server
	server := NewServer(8080, healthService)

	// Check server is correctly initialized
	assert.NotNil(t, server)
	assert.Equal(t, 8080, server.port)
	assert.Equal(t, healthService, server.healthService)
	assert.NotNil(t, server.router)
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock DB
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Create health service
	healthService := NewHealthService(sqlxDB)

	// Create server
	server := NewServer(8080, healthService)

	// Create mock delegation handler
	mockDelegationService := new(MockDelegationService)
	delegationHandler := NewDelegationHandler(mockDelegationService)

	// Setup routes
	server.SetupRoutes(delegationHandler)

	// Test health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")

	// Test health/live endpoint
	req, err = http.NewRequest("GET", "/health/live", nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")

	// Test health/ready endpoint
	req, err = http.NewRequest("GET", "/health/ready", nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")

	// Test metrics endpoint
	req, err = http.NewRequest("GET", "/metrics", nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPrepareShutdown(t *testing.T) {
	// Create a mock DB
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// Create health service
	healthService := NewHealthService(sqlxDB)

	// Create server
	server := NewServer(8080, healthService)

	// Verify initial state
	assert.False(t, healthService.isShuttingDown)

	// Prepare shutdown
	server.PrepareShutdown()

	// Verify shutdown state
	assert.True(t, healthService.isShuttingDown)
}
