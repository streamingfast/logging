package zapx

import (
	"fmt"

	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

// Trace logs a trace-level message optional fields.
//
// Example:
//
//	var zlog, tracer = logging.PackageLogger("name", "component")
//	zapx.Trace(zlog, tracer, "User logged in", zap.String("user_id", userID))
//
// This replaces the following pattern:
//
//	if tracer.Enabled() {
//	    zlog.Debug("User logged in", zap.String("user_id", userID))
//	}
//
// And should be as efficient since the check is done before any field evaluation
// and inlining should remove any overhead.
//
//go:inline
func Trace(logger *zap.Logger, tracer logging.Tracer, msg string, fields ...zap.Field) {
	if tracer.Enabled() {
		logger.Debug(msg, fields...)
	}
}

// Tracef logs a trace-level message with formatting and optional fields.
//
// Example:
//
//	var zlog, tracer = logging.PackageLogger("name", "component")
//	zapx.Tracef(zlog, tracer, "User %s logged in from %s", []any{userName, ipAddress}, zap.String("user_id", userID))
//
// This replaces the following pattern:
//
//	if tracer.Enabled() {
//	    zlog.Debug(fmt.Sprintf("User %s logged in from %s", []any{userName, ipAddress}, zap.String("user_id", userID))
//	}
//
// And should be as efficient since the check is done before any field evaluation
// and inlining should remove any overhead.
//
// Important: This is not recommended for high-frequency trace logs as the
// formatting is done eagerly.
//
//go:inline
func Tracef(logger *zap.Logger, tracer logging.Tracer, msg string, args []any, fields ...zap.Field) {
	if tracer.Enabled() {
		logger.Debug(fmt.Sprintf(msg, args...), fields...)
	}
}

// Debugf logs a debug-level message with formatting and optional fields.
//
// Example:
//
//	zapx.Debugf(zlog, "User %s logged in from %s", []any{userName, ipAddress}, zap.String("user_id", userID))
//
// Important: This is not recommended for high-frequency debug logs as the
// formatting is done eagerly.
//
//go:inline
func Debugf(logger *zap.Logger, msg string, args []any, fields ...zap.Field) {
	logger.Debug(fmt.Sprintf(msg, args...), fields...)
}

// Infof logs a info-level message with formatting and optional fields.
//
// Example:
//
//	zapx.Infof(zlog, "User %s logged in from %s", []any{userName, ipAddress}, zap.String("user_id", userID))
//
// Important: This is not recommended for high-frequency info logs as the
// formatting is done eagerly.
//
//go:inline
func Infof(logger *zap.Logger, msg string, args []any, fields ...zap.Field) {
	logger.Info(fmt.Sprintf(msg, args...), fields...)
}

// Warnf logs a warn-level message with formatting and optional fields.
//
// Example:
//
//	zapx.Warnf(zlog, "User %s logged in from %s", []any{userName, ipAddress}, zap.String("user_id", userID))
//
// Important: This is not recommended for high-frequency warn logs as the
// formatting is done eagerly.
//
//go:inline
func Warnf(logger zap.Logger, msg string, args []any, fields ...zap.Field) {
	logger.Warn(fmt.Sprintf(msg, args...), fields...)
}

// Errorf logs a error-level message with formatting and optional fields.
//
// Example:
//
//	zapx.Errorf(zlog, "User %s logged in from %s", []any{userName, ipAddress}, zap.String("user_id", userID))
//
// Important: This is not recommended for high-frequency error logs as the
// formatting is done eagerly.
//
//go:inline
func Errorf(logger *zap.Logger, msg string, args []any, fields ...zap.Field) {
	logger.Error(fmt.Sprintf(msg, args...), fields...)
}

// DPanicf logs a dpanic-level message with formatting and optional fields.
//
// Example:
//
//	zapx.DPanicf(zlog, "User %s logged in from %s", []any{userName, ipAddress}, zap.String("user_id", userID))
//
//go:inline
func DPanicf(logger *zap.Logger, msg string, args []any, fields ...zap.Field) {
	logger.DPanic(fmt.Sprintf(msg, args...), fields...)
}

// Panicf logs a panic-level message with formatting and optional fields.
//
// Example:
//
//	zapx.Panicf(zlog, "User %s logged in from %s", []any{userName, ipAddress}, zap.String("user_id", userID))
//
//go:inline
func Panicf(logger *zap.Logger, msg string, args []any, fields ...zap.Field) {
	logger.Panic(fmt.Sprintf(msg, args...), fields...)
}
