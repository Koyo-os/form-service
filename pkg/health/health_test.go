package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Koyo-os/form-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// MockHealther is a mock implementation of the Healther interface
type MockHealther struct {
	mock.Mock
}

func (m *MockHealther) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

// createTestLogger creates a logger with observer for testing
func createTestLogger() (*logger.Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	testLogger := &logger.Logger{Logger: zapLogger}
	return testLogger, recorded
}

func TestNewHealthChecker(t *testing.T) {
	testLogger, _ := createTestLogger()

	t.Run("creates health checker with no healthers", func(t *testing.T) {
		checker := NewHealthChecker(testLogger)

		assert.NotNil(t, checker)
		assert.Equal(t, testLogger, checker.logger)
		assert.Empty(t, checker.healthers)
	})

	t.Run("creates health checker with single healther", func(t *testing.T) {
		mockHealther := &MockHealther{}
		checker := NewHealthChecker(testLogger, mockHealther)

		assert.NotNil(t, checker)
		assert.Equal(t, testLogger, checker.logger)
		assert.Len(t, checker.healthers, 1)
		assert.Equal(t, mockHealther, checker.healthers[0])
	})

	t.Run("creates health checker with multiple healthers", func(t *testing.T) {
		mockHealther1 := &MockHealther{}
		mockHealther2 := &MockHealther{}
		mockHealther3 := &MockHealther{}

		checker := NewHealthChecker(testLogger, mockHealther1, mockHealther2, mockHealther3)

		assert.NotNil(t, checker)
		assert.Equal(t, testLogger, checker.logger)
		assert.Len(t, checker.healthers, 3)
		assert.Equal(t, mockHealther1, checker.healthers[0])
		assert.Equal(t, mockHealther2, checker.healthers[1])
		assert.Equal(t, mockHealther3, checker.healthers[2])
	})
}

func TestHealthChecker_HealthCheck(t *testing.T) {
	t.Run("returns OK when no healthers registered", func(t *testing.T) {
		testLogger, _ := createTestLogger()
		checker := NewHealthChecker(testLogger)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		checker.HealthCheck(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("returns OK when all healthers are healthy", func(t *testing.T) {
		testLogger, _ := createTestLogger()

		mockHealther1 := &MockHealther{}
		mockHealther1.On("IsHealthy").Return(true)

		mockHealther2 := &MockHealther{}
		mockHealther2.On("IsHealthy").Return(true)

		checker := NewHealthChecker(testLogger, mockHealther1, mockHealther2)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		checker.HealthCheck(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())

		mockHealther1.AssertExpectations(t)
		mockHealther2.AssertExpectations(t)
	})

	t.Run("returns Not OK when single healther is unhealthy", func(t *testing.T) {
		testLogger, logs := createTestLogger()

		mockHealther := &MockHealther{}
		mockHealther.On("IsHealthy").Return(false)

		checker := NewHealthChecker(testLogger, mockHealther)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		checker.HealthCheck(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Not OK", w.Body.String())

		// Verify error was logged
		assert.Equal(t, 1, logs.Len())
		assert.Equal(t, "health check failed", logs.All()[0].Message)
		assert.Equal(t, zapcore.ErrorLevel, logs.All()[0].Level)

		mockHealther.AssertExpectations(t)
	})

	t.Run("returns Not OK when any healther is unhealthy", func(t *testing.T) {
		testLogger, logs := createTestLogger()

		mockHealther1 := &MockHealther{}
		mockHealther1.On("IsHealthy").Return(true)

		mockHealther2 := &MockHealther{}
		mockHealther2.On("IsHealthy").Return(false)

		mockHealther3 := &MockHealther{}
		mockHealther3.On("IsHealthy").Return(true)

		checker := NewHealthChecker(testLogger, mockHealther1, mockHealther2, mockHealther3)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		checker.HealthCheck(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Not OK", w.Body.String())

		// Verify error was logged
		assert.Equal(t, 1, logs.Len())
		assert.Equal(t, "health check failed", logs.All()[0].Message)

		mockHealther1.AssertExpectations(t)
		mockHealther2.AssertExpectations(t)
		mockHealther3.AssertExpectations(t)
	})

	t.Run("returns Not OK when multiple healthers are unhealthy", func(t *testing.T) {
		testLogger, logs := createTestLogger()

		mockHealther1 := &MockHealther{}
		mockHealther1.On("IsHealthy").Return(false)

		mockHealther2 := &MockHealther{}
		mockHealther2.On("IsHealthy").Return(false)

		checker := NewHealthChecker(testLogger, mockHealther1, mockHealther2)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		checker.HealthCheck(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Not OK", w.Body.String())

		// Verify multiple errors were logged
		assert.Equal(t, 2, logs.Len())
		for _, logEntry := range logs.All() {
			assert.Equal(t, "health check failed", logEntry.Message)
			assert.Equal(t, zapcore.ErrorLevel, logEntry.Level)
		}

		mockHealther1.AssertExpectations(t)
		mockHealther2.AssertExpectations(t)
	})

	t.Run("checks all healthers even when some are unhealthy", func(t *testing.T) {
		testLogger, _ := createTestLogger()

		mockHealther1 := &MockHealther{}
		mockHealther1.On("IsHealthy").Return(false)

		mockHealther2 := &MockHealther{}
		mockHealther2.On("IsHealthy").Return(true)

		mockHealther3 := &MockHealther{}
		mockHealther3.On("IsHealthy").Return(false)

		checker := NewHealthChecker(testLogger, mockHealther1, mockHealther2, mockHealther3)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		checker.HealthCheck(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Not OK", w.Body.String())

		// Verify all healthers were called
		mockHealther1.AssertExpectations(t)
		mockHealther2.AssertExpectations(t)
		mockHealther3.AssertExpectations(t)
	})

	t.Run("handles HTTP request methods correctly", func(t *testing.T) {
		testLogger, _ := createTestLogger()
		checker := NewHealthChecker(testLogger)

		// Test different HTTP methods
		methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD"}

		for _, method := range methods {
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()

			checker.HealthCheck(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Method %s should return OK", method)
			if method != "HEAD" {
				assert.Equal(t, "OK", w.Body.String(), "Method %s should return OK body", method)
			}
		}
	})
}
