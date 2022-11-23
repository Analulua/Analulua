package internal

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/api/trace"
	"go.uber.org/zap"
)

// This is the default label for the correlation ID field.
const defaultCorrelationIDLabel string = "_cID"

// Errors caused by programming errors.
var (
	// ErrZapLoggerRequired indicates that the zap.Logger was not properly supplied.
	ErrZapLoggerRequired = errors.New("zap.Logger required")
)

type ZapLogger struct {
	zapLogger *zap.Logger
}

// New create new instant for the Logger.
func NewLogger(logger *zap.Logger) (*ZapLogger, error) {
	if logger == nil {
		return nil, ErrZapLoggerRequired
	}

	return &ZapLogger{
		zapLogger: logger,
	}, nil
}

func (z *ZapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zapCorrelationID(ctx))
	z.zapLogger.Info(msg, fields...)
}

func (z *ZapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zapCorrelationID(ctx))
	z.zapLogger.Warn(msg, fields...)
}

func (z *ZapLogger) Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zapCorrelationID(ctx), zap.Error(err))
	z.zapLogger.Error(msg, fields...)
}

func zapCorrelationID(ctx context.Context) zap.Field {
	ID := correlationIDFromContext(ctx)

	return zap.String(defaultCorrelationIDLabel, ID)
}

func correlationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.HasTraceID() {
		return sc.TraceID.String() + "-" + sc.SpanID.String()
	}

	return ""
}
