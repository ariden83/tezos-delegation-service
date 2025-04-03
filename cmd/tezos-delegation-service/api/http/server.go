package http

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/adapter/metrics"
	"github.com/tezos-delegation-service/internal/usecase"
)

// handlers holds the HTTP handlers.
type handlers struct {
	getDelegationHandler *GetDelegationHandler
}

// usecases holds the use case functions.
type usecases struct {
	getDelegationsFunc usecase.GetDelegationsFunc
}

// Server represents the HTTP server.
type Server struct {
	healthService *HealthService
	logger        *logrus.Entry
	metrics       metrics.Adapter
	port          uint16
	router        *gin.Engine
	handlers      *handlers
}

// NewServer creates a new HTTP server.
func NewServer(port, defaultPaginationLimit uint16, dbAdapter database.Adapter, metricClient metrics.Adapter, logger *logrus.Entry) *Server {
	u := &usecases{
		getDelegationsFunc: usecase.NewGetDelegationsFunc(defaultPaginationLimit, dbAdapter, metricClient),
	}

	h := &handlers{
		getDelegationHandler: NewGetDelegationHandler(u.getDelegationsFunc),
	}

	return &Server{
		healthService: NewHealthService(dbAdapter),
		handlers:      h,
		logger:        logger,
		metrics:       metricClient,
		port:          port,
		router:        gin.Default(),
	}
}

// SetupRoutes sets up the API routes.
func (s *Server) SetupRoutes() *Server {
	s.router.Use(metrics.Middleware(s.metrics))

	s.router.GET("/xtz/delegations", s.handlers.getDelegationHandler.GetDelegations)

	s.router.GET("/health", s.healthService.HealthHandler)
	s.router.GET("/health/live", s.healthService.LivenessHandler)
	s.router.GET("/health/ready", s.healthService.ReadinessHandler)

	s.router.GET("/metrics", metrics.PrometheusHandler())
	return s
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.healthService.SetReady(true)
	s.logger.Infof("Tezos Delegation API Server starting on port %d...", s.port)
	return s.router.Run(fmt.Sprintf(":%d", s.port))
}

// PrepareShutdown signals that the server is preparing to shut down.
func (s *Server) PrepareShutdown() {
	s.healthService.StartShutdown()
}

// WaitForShutdown waits for a shutdown signal and initiates graceful shutdown.
func (s *Server) WaitForShutdown() {
	l := s.logger.WithField("component", "shutdown")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		l.Info("Shutting down Tezos Delegation API server...")

		s.PrepareShutdown()

		l.Info("Graceful shutdown initiated. Waiting for ongoing requests to complete...")
		time.Sleep(30 * time.Second)

		l.Info("Shutdown complete")
		os.Exit(0)
	}()
}
