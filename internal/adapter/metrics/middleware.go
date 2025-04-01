package metrics

import (
	"bytes"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ResponseWriter is a custom response writer that captures the response size.
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write intercepts the response body and writes it to the buffer.
func (w ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Middleware returns a gin middleware for collecting API metrics.
func Middleware(metricsClient Adapter) gin.HandlerFunc {
	return func(c *gin.Context) {
		responseBody := &bytes.Buffer{}
		writer := ResponseWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}
		c.Writer = writer

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = strconv.FormatInt(time.Now().UnixNano(), 36)
			c.Request.Header.Set("X-Request-ID", requestID)
			c.Header("X-Request-ID", requestID)
		}

		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		status := c.Writer.Status()
		responseSize := writer.body.Len()

		if metricsClient != nil {
			metricsClient.RecordAPIRequest(
				method,
				path,
				strconv.Itoa(status),
				duration,
				responseSize,
			)
		}
	}
}

// PrometheusHandler returns a handler for the Prometheus metrics endpoint.
func PrometheusHandler() gin.HandlerFunc {
	h := gin.WrapH(promHandler())
	return func(c *gin.Context) {
		h(c)
	}
}

// promHandler returns the Prometheus HTTP handler.
func promHandler() http.Handler {
	return promhttp.Handler()
}
