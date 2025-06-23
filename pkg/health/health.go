// Package health provides health checking functionality for services.
// It allows multiple health checkers to be registered and provides HTTP endpoints
// for health status monitoring.
package health

import (
	"net/http"

	"github.com/Koyo-os/form-service/pkg/logger"
	"go.uber.org/zap"
)

type (
	// Healther defines the interface that any component must implement
	// to participate in health checking. Components implementing this
	// interface can report their health status to the health checker.
	Healther interface {
		// IsHealthy returns true if the component is healthy and ready to serve requests,
		// false otherwise. This method should perform quick checks to avoid
		// blocking the health check endpoint.
		IsHealthy() bool
	}

	// HealthChecker aggregates multiple Healther implementations and provides
	// a unified health check mechanism. It checks all registered health checkers
	// and reports the overall system health.
	HealthChecker struct {
		logger    *logger.Logger
		healthers []Healther // Collection of health checker implementations
	}
)

// NewHealthChecker creates and returns a new HealthChecker instance with
// the provided health checker implementations.
//
// Parameters:
//   - healthers: Variable number of Healther implementations to monitor
//
// Returns:
//   - *HealthChecker: Initialized health checker instance
//
// Example:
//
//	dbHealther := &DatabaseHealther{}
//	redisHealther := &RedisHealther{}
//	checker := NewHealthChecker(dbHealther, redisHealther)
func NewHealthChecker(logger *logger.Logger, healthers ...Healther) *HealthChecker {
	return &HealthChecker{
		healthers: healthers,
		logger:    logger,
	}
}

// HealthCheck is an HTTP handler that performs health checks on all registered
// health checkers and returns the overall system health status.
//
// The handler returns:
//   - HTTP 200 OK with "OK" body if all health checkers report healthy status
//   - HTTP 500 Internal Server Error with "Not OK" body if any health checker reports unhealthy status
//
// This method iterates through all registered health checkers and stops checking
// once the first unhealthy component is found for performance optimization.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request (not used but required for http.HandlerFunc signature)
func (h *HealthChecker) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ok := true

	// Check all registered health checkers
	for _, healther := range h.healthers {
		if !healther.IsHealthy() {
			ok = false
			h.logger.Error("health check failed")
		}
	}

	// Set response based on overall health status
	if ok {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Not OK"))
	}
}

// StartHealthCheckServer starts a dedicated HTTP server for health check endpoints.
// This function blocks and should typically be run in a separate goroutine.
//
// The server exposes a single endpoint:
//   - GET /health - Returns the health status of all registered components
//
// Parameters:
//   - port: The port to listen on (e.g., ":8080" or ":8081")
//   - healthChecker: The HealthChecker instance to use for health checks
//
// Example:
//
//	checker := NewHealthChecker(dbHealther, redisHealther)
//	go StartHealthCheckServer(":8081", checker)
//
// Note: This function uses the default HTTP server mux. If you need more control
// over the server configuration, consider using http.Server directly.
func (h *HealthChecker) StartHealthCheckServer(port string) {
	http.HandleFunc("/health", h.HealthCheck)
	h.logger.Info("Starting health check server", zap.String("port", port))

	if err := http.ListenAndServe(port, nil); err != nil {
		h.logger.Error("Failed to start health check server", zap.Error(err))
	}
}
