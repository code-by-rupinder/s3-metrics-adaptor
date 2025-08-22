package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
}

// LogContext holds contextual information for logging
type LogContext struct {
	Component   string
	Operation   string
	TraceID     string
	ExtraFields map[string]interface{}
}

// InitLogger initializes the logger with the specified log level and output
func InitLogger(level string, output io.Writer, prettyPrint bool, components map[string]string) error {
	// Set output
	if output == nil {
		output = os.Stdout
	}
	log.SetOutput(output)

	// Set default log level with validation
	defaultLvl, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", level, err)
	}
	log.SetLevel(defaultLvl)

	// Set log format to JSON without caller information
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		PrettyPrint:     prettyPrint,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
		// Don't use DataKey to ensure fields are at the root level
	})

	// Set up component-specific loggers
	for component, levelStr := range components {
		lvl, err := logrus.ParseLevel(levelStr)
		if err != nil {
			return fmt.Errorf("invalid log level '%s' for component '%s': %w", levelStr, component, err)
		}
		componentLogger := log.WithField("component", component)
		componentLogger.Logger.SetLevel(lvl)
	}

	// Disable caller information
	log.SetReportCaller(false)

	return nil
}

// GetLogger returns the configured logger instance
func GetLogger() *logrus.Logger {
	return log
}

// WithContext adds fields from the context to the logger entry
func WithLogContext(ctx LogContext) *logrus.Entry {
	// Create a new entry with all fields
	fields := logrus.Fields{
		"component": ctx.Component,
		"operation": ctx.Operation,
		"trace_id":  ctx.TraceID,
	}

	// Add any extra fields if present
	if ctx.ExtraFields != nil {
		for k, v := range ctx.ExtraFields {
			fields[k] = v
		}
	}

	return log.WithFields(fields)
}

// GetComponentLogger returns a logger for the specified component
func GetComponentLogger(component string) *logrus.Entry {
	return log.WithField("component", component)
}

// LogOperation logs the start and end of an operation with timing
func LogOperation(ctx LogContext, fn func() error) error {
	startTime := time.Now()

	// Log operation start in debug (detailed logging)
	WithLogContext(ctx).Debug("Starting operation")

	// Execute the operation
	err := fn()
	duration := time.Since(startTime)

	// Add duration to extra fields
	if ctx.ExtraFields == nil {
		ctx.ExtraFields = make(map[string]interface{})
	}
	ctx.ExtraFields["duration_ms"] = duration.Milliseconds()
	logger := WithLogContext(ctx)

	if err != nil {
		// Log operation failure as error
		logger.WithError(err).Error("Operation failed")
		return err
	}

	// Log operation success as debug (unless it's a significant milestone)
	logger.Debug("Operation completed successfully")
	return nil
}

// Debug logs a debug level message with context
func Debug(ctx LogContext, args ...interface{}) {
	WithLogContext(ctx).Debug(args...)
}

// Debugf logs a debug level formatted message with context
func Debugf(ctx LogContext, format string, args ...interface{}) {
	WithLogContext(ctx).Debugf(format, args...)
}

// Info logs an info level message with context
func Info(ctx LogContext, args ...interface{}) {
	WithLogContext(ctx).Info(args...)
}

// Infof logs an info level formatted message with context
func Infof(ctx LogContext, format string, args ...interface{}) {
	WithLogContext(ctx).Infof(format, args...)
}

// Warn logs a warning level message with context
func Warn(ctx LogContext, args ...interface{}) {
	WithLogContext(ctx).Warn(args...)
}

// Warnf logs a warning level formatted message with context
func Warnf(ctx LogContext, format string, args ...interface{}) {
	WithLogContext(ctx).Warnf(format, args...)
}

// Error logs an error level message with context
func Error(ctx LogContext, err error, args ...interface{}) {
	WithLogContext(ctx).WithError(err).Error(args...)
}

// Errorf logs an error level formatted message with context
func Errorf(ctx LogContext, err error, format string, args ...interface{}) {
	WithLogContext(ctx).WithError(err).Errorf(format, args...)
}

// Fatal logs a fatal level message with context and exits
func Fatal(ctx LogContext, err error, args ...interface{}) {
	WithLogContext(ctx).WithError(err).Fatal(args...)
}

// Fatalf logs a fatal level formatted message with context and exits
func Fatalf(ctx LogContext, err error, format string, args ...interface{}) {
	WithLogContext(ctx).WithError(err).Fatalf(format, args...)
}
