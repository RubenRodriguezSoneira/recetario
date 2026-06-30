package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	LoggerKey        contextKey = "logger"
)

type Logger struct {
	*slog.Logger
}

func New() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

func (l *Logger) WithCorrelationID(correlationID string) *Logger {
	return &Logger{
		Logger: l.With("correlation_id", correlationID),
	}
}

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return logger
	}
	return New()
}

func WithCorrelationID(ctx context.Context) context.Context {
	correlationID := uuid.New().String()
	logger := FromContext(ctx).WithCorrelationID(correlationID)
	ctx = WithLogger(ctx, logger)
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

func LogRequest(ctx context.Context, method, path string, duration time.Duration, statusCode int) {
	logger := FromContext(ctx)
	logger.Info("HTTP Request",
		"method", method,
		"path", path,
		"duration_ms", duration.Milliseconds(),
		"status_code", statusCode,
	)
}

func LogError(ctx context.Context, err error, msg string) {
	logger := FromContext(ctx)
	errorValue := ""
	if err != nil {
		errorValue = err.Error()
	}
	logger.Error(msg,
		"error", errorValue,
		"error_type", "application_error",
	)
}
