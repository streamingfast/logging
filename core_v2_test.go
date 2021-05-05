package logging

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestApplicationLoggerOnly(t *testing.T) {
	registry := newRegistry()
	logger := noopLogger()
	tracer := applicationLogger(registry, noEnv, "test", "com/test", &logger)

	assertLevelEnabled(t, logger, zap.InfoLevel)
	assert.False(t, tracer.Enabled())
}

func TestApplicationLoggerOnly_DebugTrue(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "true",
	})

	registry := newRegistry()
	logger := noopLogger()
	tracer := applicationLogger(registry, env, "test", "com/test", &logger)

	assertLevelEnabled(t, logger, zap.DebugLevel)
	assert.False(t, tracer.Enabled())
}

func TestApplicationLoggerOnly_DebugStar(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry()
	logger := noopLogger()
	tracer := applicationLogger(registry, env, "test", "com/test", &logger)

	assertLevelEnabled(t, logger, zap.DebugLevel)
	assert.False(t, tracer.Enabled())
}

func TestApplicationLoggerOnly_TraceStar(t *testing.T) {
	env := fakeEnv(map[string]string{
		"TRACE": "*",
	})

	registry := newRegistry()
	logger := noopLogger()
	tracer := applicationLogger(registry, env, "test", "com/test", &logger)

	assertLevelEnabled(t, logger, zap.DebugLevel)
	assert.True(t, tracer.Enabled())
}

func TestAppAndLibLogger(t *testing.T) {
	env := noEnv

	registry := newRegistry()
	libLogger := noopLogger()
	appLogger := noopLogger()

	libTracer := libraryLogger(registry, "lib", "com/lib", &libLogger)
	appTracer := applicationLogger(registry, env, "test", "com/test", &appLogger)

	assertLevelEnabled(t, libLogger, zap.PanicLevel)
	assert.False(t, libTracer.Enabled())

	assertLevelEnabled(t, appLogger, zap.InfoLevel)
	assert.False(t, appTracer.Enabled())
}

func TestAppAndLibLogger_DebugTrue(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry()
	libLogger := noopLogger()
	appLogger := noopLogger()

	libTracer := libraryLogger(registry, "lib", "com/lib", &libLogger)
	appTracer := applicationLogger(registry, env, "test", "com/test", &appLogger)

	fmt.Printf("Lib logger **%p, actual %p\n", &libLogger, libLogger)

	assertLevelEnabled(t, libLogger, zap.DebugLevel)
	assert.False(t, libTracer.Enabled())

	assertLevelEnabled(t, appLogger, zap.DebugLevel)
	assert.False(t, appTracer.Enabled())
}

func assertLevelEnabled(t *testing.T, logger *zap.Logger, level zapcore.Level) {
	t.Helper()

	shouldBeEnabled := false
	for _, candidate := range []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel} {
		if shouldBeEnabled == false && candidate == level {
			shouldBeEnabled = true
		}

		if shouldBeEnabled {
			assert.NotNil(t, logger.Check(candidate, ""), "The logger should have level %s enabled but it was not", candidate)
		} else {
			assert.Nil(t, logger.Check(candidate, ""), "The logger should have level %s disabled but it was not", candidate)
		}
	}
}

var noEnv = func(string) string { return "" }

var fakeEnv = func(in map[string]string) func(string) string {
	return func(s string) string {
		return in[s]
	}
}

func noopLogger() *zap.Logger {
	return zap.NewNop()
}
