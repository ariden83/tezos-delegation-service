package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	databasemock "github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
)

func setupHealthTestRouter(healthService *HealthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/health", healthService.HealthHandler)
	router.GET("/health/live", healthService.LivenessHandler)
	router.GET("/health/ready", healthService.ReadinessHandler)

	return router
}

func performRequest(router *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func Test_LivenessHandler(t *testing.T) {
	mockDB := databasemock.New()
	healthService := &HealthService{
		db:        mockDB,
		startTime: time.Now().Add(-time.Hour),
	}
	router := setupHealthTestRouter(healthService)

	mockDB.On("Ping").Return(nil).Once()
	w := performRequest(router, "GET", "/health/live")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alive")

	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))

	healthService.StartShutdown()
	w = performRequest(router, "GET", "/health/live")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "shutting_down")
}

func Test_ReadinessHandler(t *testing.T) {
	mockDB := databasemock.New()
	healthService := &HealthService{
		db:        mockDB,
		startTime: time.Now(),
	}
	router := setupHealthTestRouter(healthService)

	w := performRequest(router, "GET", "/health/ready")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "not_ready")

	healthService.SetReady(true)
	mockDB.On("Ping").Return(nil).Once()
	w = performRequest(router, "GET", "/health/ready")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ready")

	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))

	mockDB.On("Ping").Return(assert.AnError).Once()
	w = performRequest(router, "GET", "/health/ready")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "database_error")

	healthService.StartShutdown()
	w = performRequest(router, "GET", "/health/ready")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "shutting_down")
}

func Test_HealthHandler(t *testing.T) {
	mockDB := databasemock.New()
	healthService := &HealthService{
		db:        mockDB,
		startTime: time.Now(),
	}
	router := setupHealthTestRouter(healthService)

	mockDB.On("Ping").Return(nil).Once()
	w := performRequest(router, "GET", "/health")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")

	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))

	mockDB.On("Ping").Return(assert.AnError).Once()
	w = performRequest(router, "GET", "/health")
	assert.Equal(t, http.StatusOK, w.Code) // Still returns 200, but with degraded status
	assert.Contains(t, w.Body.String(), "error")
}

func Test_SetReadyAndIsReady(t *testing.T) {
	healthService := &HealthService{}

	assert.False(t, healthService.IsReady())

	healthService.SetReady(true)
	assert.True(t, healthService.IsReady())

	healthService.SetReady(false)
	assert.False(t, healthService.IsReady())
}

func Test_StartShutdownAndIsShuttingDown(t *testing.T) {
	healthService := &HealthService{}

	assert.False(t, healthService.IsShuttingDown())

	healthService.StartShutdown()
	assert.True(t, healthService.IsShuttingDown())
}
