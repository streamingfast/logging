package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assertLevelEnabled(t, libLogger, zap.DebugLevel)
	assert.False(t, libTracer.Enabled())

	assertLevelEnabled(t, appLogger, zap.DebugLevel)
	assert.False(t, appTracer.Enabled())
}

func TestAppAndLibLogger_LibViaLegacyRegister(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry()
	libLogger := noopLogger()
	appLogger := noopLogger()

	register(registry, "com/lib", &libLogger)
	applicationLogger(registry, env, "test", "com/test", &appLogger)

	assertLevelEnabled(t, libLogger, zap.DebugLevel)
	assertLevelEnabled(t, appLogger, zap.DebugLevel)
}

func TestLogger_CustomizedNamePerLogger(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry()
	libLogger := noopLogger()
	appLogger := noopLogger()

	testingCore := newTestingCore()

	libraryLogger(registry, "libName", "com/lib", &libLogger)
	applicationLogger(registry, env, "appName", "com/test", &appLogger, withTestingCore(testingCore))

	// Write log statements
	libLogger.Info("lib")
	appLogger.Info("app")

	require.Len(t, testingCore.checkedEntries, 2)
	assert.Equal(t, "libName", testingCore.at(0).LoggerName)
	assert.Equal(t, "appName", testingCore.at(1).LoggerName)
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

type testingCore struct {
	checkedEntries []zapcore.Entry
}

func (*testingCore) Enabled(zapcore.Level) bool          { return true }
func (c *testingCore) With([]zapcore.Field) zapcore.Core { return c }
func (c *testingCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	c.checkedEntries = append(c.checkedEntries, entry)
	return ce
}
func (c *testingCore) Write(entry zapcore.Entry, _ []zapcore.Field) error {
	return nil
}
func (*testingCore) Sync() error { return nil }

func newTestingCore() *testingCore {
	return &testingCore{}
}

func (c *testingCore) at(index int) zapcore.Entry {
	return c.checkedEntries[index]
}

func withTestingCore(t *testingCore) LoggerOption {
	return WithZapOption(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return t
	}))
}
