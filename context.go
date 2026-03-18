// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerKeyType int

const loggerKey loggerKeyType = iota

// WithLogger is used to create a new context with a logger added to it
// so it can be later retrieved using [LoggerFromContext].
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext retrieves the logger stored in the context by [WithLogger].
// If no logger is present or ctx is nil, fallbackLogger is returned instead.
//
// For one-off log calls, prefer the level-specific helpers which avoid the
// intermediate variable:
//
//	logging.Info(ctx, zlog, "user created", zap.String("id", id))
func LoggerFromContext(ctx context.Context, fallbackLogger *zap.Logger) *zap.Logger {
	if ctx == nil {
		return fallbackLogger
	}

	if ctxLogger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return ctxLogger
	}

	return fallbackLogger
}

// Logger retrieves the logger stored in the context by [WithLogger].
//
// Deprecated: Use [LoggerFromContext] instead. For one-off log calls at a
// known level you can also use the level helpers directly:
//
//	logging.Info(ctx, zlog, "msg", fields...)
//	logging.Debug(ctx, zlog, "msg", fields...)
func Logger(ctx context.Context, fallbackLogger *zap.Logger) *zap.Logger {
	return LoggerFromContext(ctx, fallbackLogger)
}

// Debug is a one-line shortcut for [LoggerFromContext](ctx, zlog).Debug(...)
func Debug(ctx context.Context, fallbackLogger *zap.Logger, msg string, fields ...zapcore.Field) {
	log(ctx, fallbackLogger, zapcore.DebugLevel, msg, fields)
}

// Info is a one-line shortcut for [LoggerFromContext](ctx, zlog).Info(...)
func Info(ctx context.Context, fallbackLogger *zap.Logger, msg string, fields ...zapcore.Field) {
	log(ctx, fallbackLogger, zapcore.InfoLevel, msg, fields)
}

// Warn is a one-line shortcut for [LoggerFromContext](ctx, zlog).Warn(...)
func Warn(ctx context.Context, fallbackLogger *zap.Logger, msg string, fields ...zapcore.Field) {
	log(ctx, fallbackLogger, zapcore.WarnLevel, msg, fields)
}

// Error is a one-line shortcut for [LoggerFromContext](ctx, zlog).Error(...)
func Error(ctx context.Context, fallbackLogger *zap.Logger, msg string, fields ...zapcore.Field) {
	log(ctx, fallbackLogger, zapcore.ErrorLevel, msg, fields)
}

func log(ctx context.Context, fallbackLogger *zap.Logger, level zapcore.Level, msg string, fields []zapcore.Field) {
	LoggerFromContext(ctx, fallbackLogger).Check(level, msg).Write(fields...)
}
