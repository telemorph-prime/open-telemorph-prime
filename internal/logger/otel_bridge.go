package logger

import (
	"context"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapOtelLogger bridges Zap logger to OpenTelemetry log.Logger interface
type zapOtelLogger struct {
	embedded.Logger
	logger *zap.Logger
}

// NewZapOtelLogger creates a new OpenTelemetry logger bridge from a Zap logger
func NewZapOtelLogger(zapLogger *zap.Logger) log.Logger {
	return &zapOtelLogger{logger: zapLogger}
}

// Enabled returns whether the logger is enabled for the given context and severity
func (z *zapOtelLogger) Enabled(ctx context.Context, params log.EnabledParameters) bool {
	// Map OpenTelemetry severity to Zap level
	level := mapSeverityToZapLevel(params.Severity)
	return z.logger.Core().Enabled(level)
}

// Emit emits a log record
func (z *zapOtelLogger) Emit(ctx context.Context, record log.Record) {
	// Convert OpenTelemetry record to Zap fields
	fields := make([]zap.Field, 0)

	// Add message
	if msg := record.Body().AsString(); msg != "" {
		fields = append(fields, zap.String("message", msg))
	}

	// Add severity
	fields = append(fields, zap.String("severity", record.Severity().String()))

	// Add attributes
	record.WalkAttributes(func(kv log.KeyValue) bool {
		field := convertKeyValueToZapField(kv.Key, kv.Value)
		fields = append(fields, field)
		return true
	})

	// Add trace context if available
	// Note: OpenTelemetry log API doesn't directly provide span context
	// Trace context should be added via attributes in the record

	// Log with appropriate level
	level := mapSeverityToZapLevel(record.Severity())
	switch level {
	case zapcore.DebugLevel:
		z.logger.Debug("", fields...)
	case zapcore.InfoLevel:
		z.logger.Info("", fields...)
	case zapcore.WarnLevel:
		z.logger.Warn("", fields...)
	case zapcore.ErrorLevel:
		z.logger.Error("", fields...)
	default:
		z.logger.Info("", fields...)
	}
}

// mapSeverityToZapLevel maps OpenTelemetry severity to Zap level
func mapSeverityToZapLevel(severity log.Severity) zapcore.Level {
	switch severity {
	case log.SeverityTrace, log.SeverityTrace2, log.SeverityTrace3, log.SeverityTrace4:
		return zapcore.DebugLevel
	case log.SeverityDebug, log.SeverityDebug2, log.SeverityDebug3, log.SeverityDebug4:
		return zapcore.DebugLevel
	case log.SeverityInfo, log.SeverityInfo2, log.SeverityInfo3, log.SeverityInfo4:
		return zapcore.InfoLevel
	case log.SeverityWarn, log.SeverityWarn2, log.SeverityWarn3, log.SeverityWarn4:
		return zapcore.WarnLevel
	case log.SeverityError, log.SeverityError2, log.SeverityError3, log.SeverityError4:
		return zapcore.ErrorLevel
	case log.SeverityFatal, log.SeverityFatal2, log.SeverityFatal3, log.SeverityFatal4:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// convertKeyValueToZapField converts an OpenTelemetry KeyValue to a Zap field
func convertKeyValueToZapField(key string, value log.Value) zap.Field {
	switch value.Kind() {
	case log.KindBool:
		return zap.Bool(key, value.AsBool())
	case log.KindInt64:
		return zap.Int64(key, value.AsInt64())
	case log.KindFloat64:
		return zap.Float64(key, value.AsFloat64())
	case log.KindString:
		return zap.String(key, value.AsString())
	case log.KindBytes:
		return zap.ByteString(key, value.AsBytes())
	case log.KindSlice:
		return zap.Any(key, value.AsSlice())
	case log.KindMap:
		return zap.Any(key, value.AsMap())
	default:
		return zap.String(key, value.AsString())
	}
}
