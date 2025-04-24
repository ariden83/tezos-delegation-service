package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	databaseadapter "github.com/tezos-delegation-service/internal/adapter/database"
)

// HealthService manages the health check functionality.
type HealthService struct {
	db              databaseadapter.Adapter
	ready           bool
	readyMu         sync.RWMutex
	startTime       time.Time
	shutdownStarted bool
	shutdownMu      sync.RWMutex
}

// NewHealthService creates a new health service.
func NewHealthService(db databaseadapter.Adapter) *HealthService {
	return &HealthService{
		db:        db,
		startTime: time.Now(),
		ready:     false,
	}
}

// SetReady sets the readiness state of the service.
func (h *HealthService) SetReady(ready bool) {
	h.readyMu.Lock()
	defer h.readyMu.Unlock()
	h.ready = ready
}

// IsReady returns whether the service is ready.
func (h *HealthService) IsReady() bool {
	h.readyMu.RLock()
	defer h.readyMu.RUnlock()
	return h.ready
}

// StartShutdown signals that the service has started shutting down.
func (h *HealthService) StartShutdown() {
	h.shutdownMu.Lock()
	defer h.shutdownMu.Unlock()
	h.shutdownStarted = true
}

// IsShuttingDown returns whether the service is shutting down.
func (h *HealthService) IsShuttingDown() bool {
	h.shutdownMu.RLock()
	defer h.shutdownMu.RUnlock()
	return h.shutdownStarted
}

// LivenessHandler handles liveness probe requests.
// The liveness probe determines if the service is running.
// It returns:
// - 200 OK if the service is alive and not shutting down.
// - 503 Service Unavailable if the service is shutting down.
func (h *HealthService) LivenessHandler(c *gin.Context) {
	if h.IsShuttingDown() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "shutting_down",
			"message": "Service is shutting down",
			"uptime":  time.Since(h.startTime).String(),
		})
		return
	}

	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, gin.H{
		"status":  "alive",
		"uptime":  time.Since(h.startTime).String(),
		"started": h.startTime,
	})
}

// ReadinessHandler handles readiness probe requests.
// The readiness probe determines if the service is ready to accept requests.
// It returns:
// - 200 OK if the service is ready and the database connection is working.
// - 503 Service Unavailable if the service is not ready or the database connection is not working.
func (h *HealthService) ReadinessHandler(c *gin.Context) {
	if !h.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not_ready",
			"message": "Service is starting up",
		})
		return
	}

	if h.IsShuttingDown() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "shutting_down",
			"message": "Service is shutting down",
		})
		return
	}

	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "database_error",
			"message": "Database connection failed",
			"error":   err.Error(),
		})
		return
	}

	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// HealthHandler handles general health check requests.
// This is a simple health check that can be used for basic monitoring.
func (h *HealthService) HealthHandler(c *gin.Context) {
	dbStatus := "ok"
	if err := h.db.Ping(); err != nil {
		dbStatus = "error"
	}

	status := "ok"
	if !h.IsReady() || h.IsShuttingDown() {
		status = "degraded"
	}

	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"uptime":   time.Since(h.startTime).String(),
		"database": dbStatus,
		"ready":    h.IsReady(),
		"shutdown": h.IsShuttingDown(),
	})
}
