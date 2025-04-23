package http

import (
	"context"
	"fmt"
	"net/http/pprof"
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
	getDelegationsHandler *GetDelegationsHandler
	getOperationsHandler  *GetOperationsHandler
	getRewardsHandler     *GetRewardsHandler
}

// usecases holds the use case functions.
type usecases struct {
	getDelegationsFunc usecase.GetDelegationsFunc
	getOperationsFunc  usecase.GetOperationsFunc
	getRewardsFunc     usecase.GetRewardsFunc
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
		getOperationsFunc:  usecase.NewGetOperationsFunc(defaultPaginationLimit, dbAdapter, metricClient),
		getRewardsFunc:     usecase.NewGetRewardsFunc(defaultPaginationLimit, dbAdapter, metricClient),
	}

	h := &handlers{
		getDelegationsHandler: NewGetDelegationsHandler(defaultPaginationLimit, u.getDelegationsFunc),
		getOperationsHandler:  NewGetOperationsHandler(defaultPaginationLimit, u.getOperationsFunc),
		getRewardsHandler:     NewGetRewardsHandler(defaultPaginationLimit, u.getRewardsFunc),
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

// corsMiddleware sets up CORS headers for the HTTP server.
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max-Delegation-ID, X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SetupRoutes sets up the API routes.
func (s *Server) SetupRoutes() *Server {
	s.router.Use(s.corsMiddleware())

	s.router.Use(metrics.Middleware(s.metrics))

	xtzGroup := s.router.Group("/xtz")
	{
		xtzGroup.GET("/delegations", s.handlers.getDelegationsHandler.GetDelegations)
		xtzGroup.GET("/operations", s.handlers.getOperationsHandler.GetOperations)
		xtzGroup.GET("/rewards", s.handlers.getRewardsHandler.GetRewards)
	}

	healthGroup := s.router.Group("/health")
	{
		healthGroup.GET("", s.healthService.HealthHandler)
		healthGroup.GET("/live", s.healthService.LivenessHandler)
		healthGroup.GET("/ready", s.healthService.ReadinessHandler)
	}

	debugGroup := s.router.Group("/debug/pprof")
	{
		debugGroup.GET("/", gin.WrapF(pprof.Index))
		debugGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		debugGroup.GET("/profile", gin.WrapF(pprof.Profile))
		debugGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		debugGroup.GET("/trace", gin.WrapF(pprof.Trace))
		debugGroup.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
		debugGroup.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
		debugGroup.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
		debugGroup.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
		debugGroup.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
		debugGroup.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
	}

	s.router.GET("/metrics", metrics.PrometheusHandler())
	return s
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.healthService.SetReady(true)
	s.logger.Infof("Tezos Delegation API Server starting on port %d...", s.port)
	return s.router.Run(fmt.Sprintf(":%d", s.port))
}

// WaitForShutdown waits for a shutdown signal and initiates graceful shutdown.
func (s *Server) WaitForShutdown(cancelFunc context.CancelFunc) {
	l := s.logger.WithField("component", "shutdown")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down Tezos Delegation service server...")
	cancelFunc()
	s.healthService.StartShutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	l.Info("Waiting for ongoing requests to complete...")

	<-ctx.Done()

	l.Info("Shutdown complete")
}
