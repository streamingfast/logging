package logging

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestApplicationLoggerOnly(t *testing.T) {
	registry := newRegistry("test", dbgZlog)
	logger, tracer := applicationLogger(registry, noEnv, "test", "com/test")

	assertLevelAndTraceEnabled(t, logger, zap.InfoLevel, tracer, traceShouldBeDisabled)
}

func TestApplicationLoggerOnly_DebugLevelOnInitialize(t *testing.T) {
	registry := newRegistry("test", dbgZlog)
	logger, tracer := applicationLogger(registry, noEnv, "test", "com/test", WithDefaultLevel(zap.DebugLevel))

	assertLevelAndTraceEnabled(t, logger, zap.DebugLevel, tracer, traceShouldBeDisabled)
}

func TestApplicationLoggerOnly_EnvDebugTrue(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "true",
	})

	registry := newRegistry("test", dbgZlog)
	logger, tracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, logger, zap.DebugLevel)
	assert.False(t, tracer.Enabled())
}

func TestApplicationLoggerOnly_EnvDebugStar(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry("test", dbgZlog)
	logger, tracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, logger, zap.DebugLevel, tracer, traceShouldBeDisabled)
}

func TestApplicationLoggerOnly_EnvDebugMatchAllRegexp(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": ".*",
	})

	registry := newRegistry("test", dbgZlog)
	logger, tracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, logger, zap.DebugLevel, tracer, traceShouldBeDisabled)
}

func TestApplicationLoggerOnly_EnvTraceStar(t *testing.T) {
	env := fakeEnv(map[string]string{
		"TRACE": "*",
	})

	registry := newRegistry("test", dbgZlog)
	logger, tracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, logger, zap.DebugLevel, tracer, traceShouldBeEnabled)
}

func TestApplicationLoggerOnly_EnvTraceMatchAllRegexp(t *testing.T) {
	env := fakeEnv(map[string]string{
		"TRACE": ".*",
	})

	registry := newRegistry("test", dbgZlog)
	logger, tracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, logger, zap.DebugLevel, tracer, traceShouldBeEnabled)
}

func TestAppAndPkgLogger(t *testing.T) {
	env := noEnv

	registry := newRegistry("test", dbgZlog)

	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, pkgLogger, zap.ErrorLevel, pkgTracer, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.InfoLevel, appTracer, traceShouldBeDisabled)
}

func TestTree_InitializeSpecOneRegexMatching2Packages(t *testing.T) {
	env := noEnv

	registry := newRegistry("test", dbgZlog)

	pkgLogger1, pkgTracer1 := packageLogger(registry, "lib1", "com/lib/1")
	pkgLogger2, pkgTracer2 := packageLogger(registry, "lib2", "com/lib/2")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test", WithDefaultSpec(
		"com/lib*=debug",
	))

	assertLevelAndTraceEnabled(t, pkgLogger1, zap.DebugLevel, pkgTracer1, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, pkgLogger2, zap.DebugLevel, pkgTracer2, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.InfoLevel, appTracer, traceShouldBeDisabled)
}

func TestAppAndPkgLogger_SameShortNameStartsAllInInfo(t *testing.T) {
	env := noEnv

	registry := newRegistry("test", dbgZlog)
	pkgLogger, pkgTracer := packageLogger(registry, "test", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, pkgLogger, zap.InfoLevel, pkgTracer, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.InfoLevel, appTracer, traceShouldBeDisabled)
}

func TestAppAndPkgLogger_EnvDebugTrue(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry("test", dbgZlog)
	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, pkgLogger, zap.DebugLevel, pkgTracer, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.DebugLevel, appTracer, traceShouldBeDisabled)
}

func TestAppAndPkgLogger_EnvDebugSpecificPackage(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "com/test",
	})

	registry := newRegistry("test", dbgZlog)
	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, pkgLogger, zap.ErrorLevel, pkgTracer, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.DebugLevel, appTracer, traceShouldBeDisabled)
}

func TestAppAndPkgLogger_EnvDebugSpecificPackageRegex(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "com/(test|lib)",
	})

	registry := newRegistry("test", dbgZlog)
	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, pkgLogger, zap.DebugLevel, pkgTracer, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.DebugLevel, appTracer, traceShouldBeDisabled)
}

func TestAppAndPkgLogger_EnvDebugAllPkgViaLegacyRegister(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry("test", dbgZlog)
	pkgLogger := zap.NewNop()

	register(registry, "com/lib", pkgLogger)
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelEnabled(t, pkgLogger, zap.DebugLevel)
	assertLevelAndTraceEnabled(t, appLogger, zap.DebugLevel, appTracer, traceShouldBeDisabled)
}

func TestAppAndPkgLogger_LoggerDefaultLevel(t *testing.T) {
	registry := newRegistry("test", dbgZlog)
	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib", LoggerDefaultLevel(zap.DebugLevel))
	appLogger, appTracer := applicationLogger(registry, noEnv, "test", "com/test")

	assertLevelAndTraceEnabled(t, pkgLogger, zap.DebugLevel, pkgTracer, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.InfoLevel, appTracer, traceShouldBeDisabled)
}

func TestAppAndPkgLogger_LoggerDefaultLevel_EnvDlogOverrideLibToDebug(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DLOG": "com/lib=debug",
	})

	registry := newRegistry("test", dbgZlog)
	pkgLogger, pkgTracer := packageLogger(registry, "lib", "com/lib", LoggerDefaultLevel(zap.InfoLevel))
	appLogger, appTracer := applicationLogger(registry, env, "test", "com/test")

	assertLevelAndTraceEnabled(t, pkgLogger, zap.DebugLevel, pkgTracer, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.InfoLevel, appTracer, traceShouldBeDisabled)
}

func TestLogger_CustomizedNamePerLogger(t *testing.T) {
	env := fakeEnv(map[string]string{
		"DEBUG": "*",
	})

	registry := newRegistry("test", dbgZlog)
	testingCore := newTestingCore()

	pkgLogger, _ := packageLogger(registry, "libName", "com/lib")
	appLogger, _ := applicationLogger(registry, env, "appName", "com/test", testingOptions(testingCore))

	// Write log statements
	pkgLogger.Info("lib")
	appLogger.Info("app")

	require.Len(t, testingCore.checkedEntries, 2)
	assert.Equal(t, "libName", testingCore.at(0).LoggerName)
	assert.Equal(t, "appName", testingCore.at(1).LoggerName)
}

func TestLogger_WriteToLogFile(t *testing.T) {
	file, err := os.CreateTemp("", "logging-log-file.json")
	require.NoError(t, err)

	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	registry := newRegistry("test", dbgZlog)
	pkgLogger, _ := packageLogger(registry, "libName", "com/lib")
	appLogger, _ := applicationLogger(registry, noEnv, "appName", "com/test", WithOutputToFile(file.Name()))

	// Write log statements
	pkgLogger.Info("lib")
	appLogger.Info("app")

	pkgLogger.Sync()
	appLogger.Sync()

	content, err := os.ReadFile(file.Name())
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")
	assert.Equal(t, 2, len(lines))
	assert.Contains(t, lines[0], `"logger":"appName","msg":"app"`)
	assert.Equal(t, "", lines[1])
}

func TestLogger_SetLevelForEntry_EnvDebugStar(t *testing.T) {
	env := fakeEnv(map[string]string{})

	registry := newRegistry("test", dbgZlog)
	pkgLogger, pkgTracer := packageLogger(registry, "libName", "com/lib")
	appLogger, appTracer := applicationLogger(registry, env, "appName", "com/test")

	overrideEnv := fakeEnv(map[string]string{
		"DEBUG": "*",
	})
	registry.forAllEntriesMatchingSpec(newLogLevelSpec(overrideEnv), func(entry *registryEntry, level zapcore.Level, trace bool) {
		registry.setLevelForEntry(entry, level, trace)
	})

	assertLevelAndTraceEnabled(t, pkgLogger, zap.DebugLevel, pkgTracer, traceShouldBeDisabled)
	assertLevelAndTraceEnabled(t, appLogger, zap.DebugLevel, appTracer, traceShouldBeDisabled)
}

var traceShouldBeEnabled = true
var traceShouldBeDisabled = false

func assertLevelAndTraceEnabled(t *testing.T, logger *zap.Logger, expectedEnabledLevel zapcore.Level, tracer Tracer, shouldTraceEnabled bool) {
	t.Helper()

	assertLevelEnabled(t, logger, expectedEnabledLevel)
	if shouldTraceEnabled {
		assert.True(t, tracer.Enabled(), "trace is disabled for logger's tracer but we expected it to be enabled")
	} else {
		assert.False(t, tracer.Enabled(), "trace is enabled for logger's tracer but we expected it to be disabled")
	}
}

func assertLevelEnabled(t *testing.T, logger *zap.Logger, expectedEnabledLevel zapcore.Level) {
	t.Helper()

	var assertions []string
	shouldBeEnabled := false
	for _, candidate := range []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel} {
		if shouldBeEnabled == false && candidate == expectedEnabledLevel {
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
		assertions = append(assertions, computeLevelsString(logger.Core()))

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

func testingOptions(t *testingCore) InstantiateOption {
	return withZapOption(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return t
	}))
}
