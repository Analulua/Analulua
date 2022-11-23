package telemetry

import (
	"context"

	"go.uber.org/zap"
)

// Logger Interface. All methods SHOULD be safe for concurrent use.
type Logger interface {
	// Info logs a message at Info level
	Info(ctx context.Context, msg string, fields ...zap.Field)
	// Warn logs a message at Warn level
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	// Error logs a message at Errors level
	Error(ctx context.Context, msg string, err error, fields ...zap.Field)
}
