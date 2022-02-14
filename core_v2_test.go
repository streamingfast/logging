package logging

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestApplicationLoggerOnly(t *testing.T) {
	registry := newRegistry("test")
	logger, tracer := applicationLogger(registry, noEnv, "test", "com/test")

	assertLevelEnabled(t, logger, zap.InfoLevel)
	assert.False(t, tracer.Enabled())
}

func TestApplicationLoggerOnly_DebugTrue(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "true",
	})

	registry := newRegistry("test")
	logger, tracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, logger, zap.DebugLevel)
	assert.False(t, tracer.Enabled())
}

func TestApplicationLoggerOnly_DebugStart(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry("test")
	logger, tracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, logger, zap.DebugLevel)
	assert.False(t, tracer.Enabled())
}

func TestApplicationLoggerOnly_TraceStart(t *testing.T) {
	env := fakeEnv(map[string]string{
		"TRACE": "*",
	})

	registry := newRegistry("test")
	logger, tracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, logger, zap.DebugLevel)
	assert.True(t, tracer.Enabled())
}

func TestAppAndPkgLogger(t *testing.T) {
	env := noEnv

	registry := newRegistry("test")

	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, pkgLogger, zap.ErrorLevel)
	assert.False(t, pkgTracer.Enabled())

	assertLevelEnabled(t, appLogger, zap.InfoLevel)
	assert.False(t, appTracer.Enabled())
}

func TestAppAndPkgLogger_SameShortNameStartsAllInInfo(t *testing.T) {
	env := noEnv

	registry := newRegistry("test")
	pkgLogger, pkgTracer := packageLogger(registry, "test", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, pkgLogger, zap.InfoLevel)
	assert.False(t, pkgTracer.Enabled())

	assertLevelEnabled(t, appLogger, zap.InfoLevel)
	assert.False(t, appTracer.Enabled())
}

func TestAppAndPkgLogger_DebugTrue(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry("test")
	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, pkgLogger, zap.DebugLevel)
	assert.False(t, pkgTracer.Enabled())

	assertLevelEnabled(t, appLogger, zap.DebugLevel)
	assert.False(t, appTracer.Enabled())
}

func TestAppAndPkgLogger_DebugSpecificPackage(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "com/test",
	})

	registry := newRegistry("test")
	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, pkgLogger, zap.ErrorLevel)
	assert.False(t, pkgTracer.Enabled())

	assertLevelEnabled(t, appLogger, zap.DebugLevel)
	assert.False(t, appTracer.Enabled())
}

func TestAppAndPkgLogger_DebugSpecificPackageRegex(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "com/(test|lib)",
	})

	registry := newRegistry("test")
	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, pkgLogger, zap.DebugLevel)
	assert.False(t, pkgTracer.Enabled())

	assertLevelEnabled(t, appLogger, zap.DebugLevel)
	assert.False(t, appTracer.Enabled())
}

func TestAppAndPkgLogger_PkgViaLegacyRegister(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry("test")
	pkgLogger := zap.NewNop()

	register(registry, "com/lib", pkgLogger)
	appLogger, _ := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, pkgLogger, zap.DebugLevel)
	assertLevelEnabled(t, appLogger, zap.DebugLevel)
}

func TestLogger_CustomizedNamePerLogger(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry("test")
	testingCore := newTestingCore()

	pkgLogger, _ := packageLogger(registry, "libName", "com/lib")
	appLogger, _ := applicationLogger(registry, env, "appName", "com/test", withTestingCore(testingCore))

	// Write log statements
	pkgLogger.Info("lib")
	appLogger.Info("app")

	require.Len(t, testingCore.checkedEntries, 2)
	assert.Equal(t, "libName", testingCore.at(0).LoggerName)
	assert.Equal(t, "appName", testingCore.at(1).LoggerName)
}

func TestLogger_SetLevelForEntry_Debug(t *testing.T) {
	env := fakeEnv(map[string]string{})

	registry := newRegistry("test")
	pkgLogger, _ := packageLogger(registry, "libName", "com/lib")
	appLogger, _ := applicationLogger(registry, env, "appName", "com/test")

	overrideEnv := fakeEnv(map[string]string{
		"DEBUG": "*",
	})
	registry.forAllEntriesMatchingSpec(newLogLevelSpec(overrideEnv), func(entry *registryEntry, level zapcore.Level, trace bool) {
		registry.setLevelForEntry(entry, level, trace)
	})

	assertLevelEnabled(t, pkgLogger, zap.DebugLevel)
	assertLevelEnabled(t, appLogger, zap.DebugLevel)
}

func assertLevelEnabled(t *testing.T, logger *zap.Logger, level zapcore.Level) {
	t.Helper()

	var assertions []string

	shouldBeEnabled := false
	for _, candidate := range []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel} {
		if shouldBeEnabled == false && candidate == level {
			shouldBeEnabled = true
		}

		if shouldBeEnabled {
			if logger.Check(candidate, "") == nil {
				assertions = append(assertions, fmt.Sprintf("The logger should have level %s enabled but it was not", candidate))
			}
		} else {
			if logger.Check(candidate, "") != nil {
				assertions = append(assertions, fmt.Sprintf("The logger should have level %s enabled but it was not", candidate))
			}
		}
	}

	if len(assertions) > 0 {
		assert.Fail(t, strings.Join(assertions, "\n"))
	}
}

var noEnv = func(string) string { return "" }

var fakeEnv = func(in map[string]string) func(string) string {
	return func(s string) string {
		return in[s]
	}
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
