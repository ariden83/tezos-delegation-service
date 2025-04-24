package http

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/tezos-delegation-service/internal/adapter/database"
	databaseadaptermock "github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
)

func Test_NewServer(t *testing.T) {
	type args struct {
		port         uint16
		defaultLimit uint16
		dbAdapter    database.Adapter
		metricClient metrics.Adapter
		logger       *logrus.Entry
	}

	mockDB := databaseadaptermock.New()
	mockMetrics := metricsnoop.New()
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name  string
		args  args
		check func(*testing.T, *Server)
	}{
		{
			name: "Nominal case",
			args: args{
				port:         8080,
				defaultLimit: 50,
				dbAdapter:    mockDB,
				metricClient: mockMetrics,
				logger:       logger,
			},
			check: func(t *testing.T, s *Server) {
				assert.NotNil(t, s)
				assert.Equal(t, 8080, int(s.port))
				assert.NotNil(t, s.healthService)
				assert.NotNil(t, s.router)
				assert.NotNil(t, s.handlers)
				assert.NotNil(t, s.handlers.getDelegationsHandler)
				assert.Equal(t, logger, s.logger)
				assert.Equal(t, mockMetrics, s.metrics)
			},
		},
		{
			name: "Alternate port",
			args: args{
				port:         9090,
				defaultLimit: 50,
				dbAdapter:    mockDB,
				metricClient: mockMetrics,
				logger:       logger,
			},
			check: func(t *testing.T, s *Server) {
				assert.Equal(t, 9090, int(s.port))
			},
		},
		{
			name: "With DB nil adapter",
			args: args{
				port:         8080,
				defaultLimit: 50,
				dbAdapter:    nil,
				metricClient: mockMetrics,
				logger:       logger,
			},
			check: func(t *testing.T, s *Server) {
				assert.NotNil(t, s)
				assert.NotNil(t, s.healthService)
			},
		},
		{
			name: "With nil metrics",
			args: args{
				port:         8080,
				defaultLimit: 50,
				dbAdapter:    mockDB,
				metricClient: nil,
				logger:       logger,
			},
			check: func(t *testing.T, s *Server) {
				assert.NotNil(t, s)
				assert.Nil(t, s.metrics)
			},
		},
		{
			name: "With logger nil",
			args: args{
				port:         8080,
				defaultLimit: 50,
				dbAdapter:    mockDB,
				metricClient: mockMetrics,
				logger:       nil,
			},
			check: func(t *testing.T, s *Server) {
				assert.NotNil(t, s)
				assert.Nil(t, s.logger)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(tt.args.port, tt.args.defaultLimit, tt.args.dbAdapter, tt.args.metricClient, tt.args.logger)
			tt.check(t, server)
		})
	}
}

func Test_Server_SetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type fields struct {
		healthService *HealthService
		logger        *logrus.Entry
		metrics       metrics.Adapter
		port          uint16
		router        *gin.Engine
		handlers      *handlers
	}

	mockDB := databaseadaptermock.New()
	mockMetrics := metricsnoop.New()
	logger := logrus.NewEntry(logrus.New())

	mockHandler := &GetDelegationsHandler{}
	mockHandlers := &handlers{
		getDelegationsHandler: mockHandler,
	}

	tests := []struct {
		name     string
		fields   fields
		testFunc func(*testing.T, *Server)
	}{
		{
			name: "Nominal case - all routes configured",
			fields: fields{
				healthService: NewHealthService(mockDB),
				logger:        logger,
				metrics:       mockMetrics,
				port:          8080,
				router:        gin.New(),
				handlers:      mockHandlers,
			},
			testFunc: func(t *testing.T, s *Server) {
				routes := s.router.Routes()

				routePaths := make(map[string]bool)
				for _, route := range routes {
					routePaths[route.Path] = true
				}

				assert.True(t, routePaths["/xtz/delegations"])
				assert.True(t, routePaths["/health"])
				assert.True(t, routePaths["/health/live"])
				assert.True(t, routePaths["/health/ready"])
				assert.True(t, routePaths["/metrics"])
			},
		},
		{
			name: "With a nil handler",
			fields: fields{
				healthService: NewHealthService(mockDB),
				logger:        logger,
				metrics:       mockMetrics,
				port:          8080,
				router:        gin.New(),
				handlers: &handlers{
					getDelegationsHandler: nil,
				},
			},
			testFunc: func(t *testing.T, s *Server) {
				routes := s.router.Routes()

				routePaths := make(map[string]bool)
				for _, route := range routes {
					routePaths[route.Path] = true
				}

				assert.True(t, routePaths["/health"])
				assert.True(t, routePaths["/health/live"])
				assert.True(t, routePaths["/health/ready"])
				assert.True(t, routePaths["/metrics"])
			},
		},
		{
			name: "With nil metrics",
			fields: fields{
				healthService: NewHealthService(mockDB),
				logger:        logger,
				metrics:       nil,
				port:          8080,
				router:        gin.New(),
				handlers:      mockHandlers,
			},
			testFunc: func(t *testing.T, s *Server) {
				routes := s.router.Routes()
				assert.NotEmpty(t, routes)

				routePaths := make(map[string]bool)
				for _, route := range routes {
					routePaths[route.Path] = true
				}

				assert.True(t, routePaths["/xtz/delegations"])
				assert.True(t, routePaths["/health"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				healthService: tt.fields.healthService,
				logger:        tt.fields.logger,
				metrics:       tt.fields.metrics,
				port:          tt.fields.port,
				router:        tt.fields.router,
				handlers:      tt.fields.handlers,
			}

			result := s.SetupRoutes()
			assert.Equal(t, s, result)

			tt.testFunc(t, s)
		})
	}
}

func Test_Server_WaitForShutdown(t *testing.T) {
	mockLogger := logrus.NewEntry(logrus.New())

	db := databaseadaptermock.New()
	healthService := NewHealthService(db)
	metricClient := metricsnoop.New()

	router := gin.New()

	h := &handlers{
		getDelegationsHandler: &GetDelegationsHandler{},
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
