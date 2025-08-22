package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testComponent      = "test-component"
	testTraceID        = "test-trace-123"
	testOperation      = "test-operation"
	testMessage        = "test message"
	testErrorMessage   = "test error"
	testCount          = 42
	testCountTemplate  = "count: %d"
	testCountFormatted = "count: 42"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		components    map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name:        "debug level",
			level:       "debug",
			components:  map[string]string{"component1": "debug"},
			expectError: false,
		},
		{
			name:        "info level",
			level:       "info",
			components:  map[string]string{"component1": "info"},
			expectError: false,
		},
		{
			name:        "error level",
			level:       "error",
			components:  map[string]string{"component1": "error"},
			expectError: false,
		},
		{
			name:          "invalid level",
			level:         "invalid",
			expectError:   true,
			errorContains: "invalid log level",
		},
		{
			name:          "invalid component level",
			level:         "info",
			components:    map[string]string{"component1": "invalid"},
			expectError:   true,
			errorContains: "invalid log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLogger(tt.level, &bytes.Buffer{}, false, tt.components)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	err := InitLogger("debug", buf, false, nil)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		logFunc     func()
		level       string
		message     string
		component   string
		operation   string
		hasError    bool
		errorString string
	}{
		{
			name: "info message",
			logFunc: func() {
				Info(LogContext{
					Component: testComponent,
					Operation: testOperation,
					TraceID:   testTraceID,
				}, testMessage)
			},
			level:     "info",
			message:   testMessage,
			component: testComponent,
			operation: testOperation,
		},
		{
			name: "error message",
			logFunc: func() {
				testErr := errors.New("test error")
				Error(LogContext{
					Component: testComponent,
					Operation: testOperation,
					TraceID:   testTraceID,
				}, testErr, testMessage)
			},
			level:       "error",
			message:     testMessage,
			component:   testComponent,
			operation:   testOperation,
			hasError:    true,
			errorString: "test error",
		},
		{
			name: "debug message",
			logFunc: func() {
				Debug(LogContext{
					Component: testComponent,
					Operation: testOperation,
					TraceID:   testTraceID,
				}, testMessage)
			},
			level:     "debug",
			message:   testMessage,
			component: testComponent,
			operation: testOperation,
		},
		{
			name: "warn message",
			logFunc: func() {
				Warn(LogContext{
					Component: testComponent,
					Operation: testOperation,
					TraceID:   testTraceID,
				}, testMessage)
			},
			level:     "warning",
			message:   testMessage,
			component: testComponent,
			operation: testOperation,
		},
		{
			name: "formatted info message",
			logFunc: func() {
				Infof(LogContext{
					Component: testComponent,
					Operation: testOperation,
					TraceID:   testTraceID,
				}, "%s: %d", "count", 42)
			},
			level:     "info",
			message:   "count: 42",
			component: testComponent,
			operation: testOperation,
		},
		{
			name: "formatted error message",
			logFunc: func() {
				testErr := errors.New("test error")
				Errorf(LogContext{
					Component: testComponent,
					Operation: testOperation,
					TraceID:   testTraceID,
				}, testErr, "%s: %d", "count", 42)
			},
			level:       "error",
			message:     "count: 42",
			component:   testComponent,
			operation:   testOperation,
			hasError:    true,
			errorString: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			assert.Equal(t, tt.message, logEntry["message"])
			assert.Equal(t, tt.level, logEntry["level"])
			assert.Equal(t, tt.component, logEntry["component"])
			assert.Equal(t, tt.operation, logEntry["operation"])
			assert.Equal(t, testTraceID, logEntry["trace_id"])

			if tt.hasError {
				assert.Equal(t, tt.errorString, logEntry["error"])
			}
		})
	}
}

func TestLogLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	components := map[string]string{
		"test-component": "info",
	}
	err := InitLogger("info", buf, false, components)
	assert.NoError(t, err)

	// This debug message should not appear since the component level is info
	Debug(LogContext{Component: "test-component"}, "Debug message")

	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	if err == nil {
		t.Error("Debug message appeared when component level is info")
	}

	// This info message should appear
	buf.Reset()
	Info(LogContext{Component: "test-component"}, "Info message")
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "Info message", logEntry["message"])
}

func TestLogOperation(t *testing.T) {
	buf := &bytes.Buffer{}
	err := InitLogger("debug", buf, false, nil)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		operation   func() error
		expectError bool
	}{
		{
			name: "successful operation",
			operation: func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			expectError: false,
		},
		{
			name: "failed operation",
			operation: func() error {
				time.Sleep(10 * time.Millisecond)
				return errors.New("operation failed")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			ctx := LogContext{
				Component: testComponent,
				Operation: testOperation,
				TraceID:   testTraceID,
			}

			err := LogOperation(ctx, tt.operation)

			var entries []map[string]interface{}
			for _, line := range strings.Split(buf.String(), "\n") {
				if line == "" {
					continue
				}
				var entry map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(line), &entry))
				entries = append(entries, entry)
			}

			// Check start message
			assert.Equal(t, "Starting operation", entries[0]["message"])
			assert.Equal(t, testComponent, entries[0]["component"])
			assert.Equal(t, testOperation, entries[0]["operation"])
			assert.Equal(t, testTraceID, entries[0]["trace_id"])

			// Check completion/error message
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, "Operation failed", entries[1]["message"])
				assert.Equal(t, "operation failed", entries[1]["error"])
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "Operation completed successfully", entries[1]["message"])
			}

			// Check duration was logged
			assert.Contains(t, entries[1], "duration_ms")
			duration := entries[1]["duration_ms"].(float64)
			assert.GreaterOrEqual(t, duration, float64(10))
		})
	}
}

func TestCustomOutput(t *testing.T) {
	customBuf := &bytes.Buffer{}
	err := InitLogger("info", customBuf, false, nil)
	assert.NoError(t, err)

	testMessage := "test message to custom output"
	Info(LogContext{Component: testComponent}, testMessage)

	var logEntry map[string]interface{}
	err = json.Unmarshal(customBuf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, testMessage, logEntry["message"])
}

func TestPrettyPrintOutput(t *testing.T) {
	tests := []struct {
		name        string
		prettyPrint bool
	}{
		{
			name:        "pretty print enabled",
			prettyPrint: true,
		},
		{
			name:        "pretty print disabled",
			prettyPrint: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := InitLogger("info", buf, tt.prettyPrint, nil)
			assert.NoError(t, err)

			Info(LogContext{Component: testComponent}, testMessage)

			output := buf.String()
			var logEntry map[string]interface{}
			err = json.Unmarshal([]byte(output), &logEntry)
			assert.NoError(t, err)

			if tt.prettyPrint {
				assert.True(t, strings.Count(output, "\n") > 0)
				assert.True(t, strings.Contains(output, "  "))
			} else {
				assert.Equal(t, 1, strings.Count(output, "\n"))
				assert.False(t, strings.Contains(output, "  "))
			}
		})
	}
}
