package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	databaseadaptermock "github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
	tzktapimock "github.com/tezos-delegation-service/internal/adapter/tzktapi/impl/mock"
)

func Test_NewServer(t *testing.T) {
	db := databaseadaptermock.New()
	tzktAPI := tzktapimock.New()
	providedPort := 8080
	metricClient := metricsnoop.New()
	logger := logrus.NewEntry(logrus.New())

	healthService := NewHealthService(db)

	server := NewServer(providedPort, tzktAPI, db, metricClient, logger)

	assert.NotNil(t, server)
	assert.Equal(t, providedPort, server.port)
	assert.Equal(t, healthService, server.healthService)
	assert.NotNil(t, server.router)
}

func Test_SetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := databaseadaptermock.New()
	tzktAPI := tzktapimock.New()
	providedPort := 8080
	metricClient := metricsnoop.New()
	logger := logrus.NewEntry(logrus.New())

	server := NewServer(providedPort, tzktAPI, db, metricClient, logger).SetupRoutes()

	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")

	req, err = http.NewRequest("GET", "/health/live", nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")

	req, err = http.NewRequest("GET", "/health/ready", nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")

	req, err = http.NewRequest("GET", "/metrics", nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_PrepareShutdown(t *testing.T) {
	db := databaseadaptermock.New()
	tzktAPI := tzktapimock.New()
	providedPort := 8080
	metricClient := metricsnoop.New()
	logger := logrus.NewEntry(logrus.New())

	server := NewServer(providedPort, tzktAPI, db, metricClient, logger).SetupRoutes()

	assert.False(t, server.healthService.shutdownStarted)

	server.PrepareShutdown()

	assert.True(t, server.healthService.shutdownStarted)
}

func Test_Server_WaitForShutdown(t *testing.T) {
	mockLogger := logrus.NewEntry(logrus.New())

	db := databaseadaptermock.New()
	healthService := NewHealthService(db)
	metricClient := metricsnoop.New()

	router := gin.New()

	h := &handlers{
		getDelegationHandler: &GetDelegationHandler{},
	}

	s := &Server{
		healthService: healthService,
		logger:        mockLogger,
		metrics:       metricClient,
		port:          8080,
		router:        router,
		handlers:      h,
	}

	done := make(chan bool)
	go func() {
		s.WaitForShutdown()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	assert.True(t, true)
}
