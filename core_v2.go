package logging

import (
	"fmt"
	"net/http"
	"os"

	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
)

var dbgZlog = zap.NewNop()
var dbgRegistry = newRegistry("logging_dbg")

func init() {
	registerDebug(dbgRegistry, "logging", "github.com/streamingfast/logging", dbgZlog)
}

type loggerFactory func(name string, level zap.AtomicLevel) *zap.Logger

// This v2 version of `core.go` is a work in progress without any backward compatibility
// version. It might not made it to an official version of the library so you can depend
// on it at your own risk.

type loggerOptions struct {
	encoderVerbosity         *int
	level                    *zap.AtomicLevel
	reportAllErrors          *bool
	serviceName              *string
	switcherServerAutoStart  *bool
	switcherServerListenAddr string
	zapOptions               []zap.Option
	registerOptions          []RegisterOption

	forceProductionLogger bool
	// Use internally only, no With... value defined for it
	loggerName string
}

func newLoggerOptions(shortName string, opts ...LoggerOption) loggerOptions {
	loggerOptions := loggerOptions{switcherServerListenAddr: "127.0.0.1:1065"}
	for _, opt := range opts {
		opt.apply(&loggerOptions)
	}

	if loggerOptions.reportAllErrors == nil {
		WithReportAllErrors().apply(&loggerOptions)
	}

	if loggerOptions.serviceName == nil && shortName != "" {
		WithServiceName(shortName).apply(&loggerOptions)
	}

	if loggerOptions.switcherServerAutoStart == nil && loggerOptions.isProductionEnvironment() {
		WithSwitcherServerAutoStart().apply(&loggerOptions)
	}

	return loggerOptions
}

type LoggerOption interface {
	apply(o *loggerOptions)
}

type loggerFuncOption func(o *loggerOptions)

func (f loggerFuncOption) apply(o *loggerOptions) {
	f(o)
}

func WithSwitcherServerAutoStart() LoggerOption {
	return loggerFuncOption(func(o *loggerOptions) {
		o.switcherServerAutoStart = ptrBool(true)
	})
}

func WithAtomicLevel(level zap.AtomicLevel) LoggerOption {
	return loggerFuncOption(func(o *loggerOptions) {
		o.level = ptrLevel(level)
	})
}

func WithReportAllErrors() LoggerOption {
	return loggerFuncOption(func(o *loggerOptions) {
		o.reportAllErrors = ptrBool(true)
	})
}

func WithServiceName(name string) LoggerOption {
	return loggerFuncOption(func(o *loggerOptions) {
		o.serviceName = ptrString(name)
	})
}

func WithZapOption(zapOption zap.Option) LoggerOption {
	return loggerFuncOption(func(o *loggerOptions) {
		o.zapOptions = append(o.zapOptions, zapOption)
	})
}

func WithProductionLogger() LoggerOption {
	return loggerFuncOption(func(o *loggerOptions) {
		o.forceProductionLogger = true
	})
}

func WithOnUpdate(onUpdate func(newLogger *zap.Logger)) LoggerOption {
	return loggerFuncOption(func(o *loggerOptions) {
		o.registerOptions = append(o.registerOptions, RegisterOnUpdate(onUpdate))
	})
}

func (o *loggerOptions) isProductionEnvironment() bool {
	if o.forceProductionLogger {
		return true
	}

	_, err := os.Stat("/.dockerenv")

	return !os.IsNotExist(err)
}

// PackageLogger creates a new no-op logger (via `zap.NewNop`) and automatically registered it
// withing the logging registry with a tracer that can be be used for conditionally tracing
// code.
//
// You should used this in packages that are not `main` packages
func PackageLogger(shortName string, packageID string, registerOptions ...RegisterOption) (*zap.Logger, Tracer) {
	return packageLogger(globalRegistry, shortName, packageID, registerOptions...)
}

func packageLogger(registry *registry, shortName string, packageID string, registerOptions ...RegisterOption) (*zap.Logger, Tracer) {
	return register2(registry, shortName, packageID, registerOptions...)
}

// ApplicationLogger should be used to get a logger for a top-level binary application which will
// immediately activate all registered loggers with a logger. The actual logger for all component
// used is deried based on the identified environment and from environment variables.
//
// Here the set of rules used and the outcome they are giving:
//
//  1. If a production environment is detected (for now, only checking if file /.dockerenv exists)
//     Use a JSON StackDriver compatible format
//
//  2. Otherwise
//     Use a developer friendly colored format
//
//
// *Note* The ApplicationLogger should be start only once per processed. That could be enforced
//        in the future.
func ApplicationLogger(shortName string, packageID string, opts ...LoggerOption) (*zap.Logger, Tracer) {
	return applicationLogger(globalRegistry, os.Getenv, shortName, packageID, opts...)
}

func applicationLogger(
	registry *registry,
	envGet func(string) string,
	shortName string,
	packageID string,
	opts ...LoggerOption,
) (*zap.Logger, Tracer) {
	loggerOptions := newLoggerOptions(shortName, opts...)
	dbgZlog.Info("application logger invoked")
	logger, tracer := register2(registry, shortName, packageID, loggerOptions.registerOptions...)

	registry.factory = func(name string, level zap.AtomicLevel) *zap.Logger {
		clonedOptions := loggerOptions
		if name != "" {
			clonedOptions.loggerName = name
		}

		// If the level was specified up-front, let's not use the one received
		if loggerOptions.level != nil {
			return newLogger(&clonedOptions)
		}

		clonedOptions.level = ptrLevel(level)

		return newLogger(&clonedOptions)
	}

	// We first initialize all logger to something at development panic (so writes
	// development panics, panics and fatals).
	//
	// This ensure that all loggers are pre-created and as such, we are able to override
	// the level of any of them (because we have created it).
	registry.forAllEntries(func(entry *registryEntry) {
		registry.setLoggerForEntry(entry, zapcore.ErrorLevel, false)
	})

	// We then override the level based on the spec extracted from the environment
	appLoggerAffectedByEnv := false
	logLevelSpec := newLogLevelSpec(envGet)
	registry.forAllEntriesMatchingSpec(logLevelSpec, func(entry *registryEntry, level zapcore.Level, trace bool) {
		if entry.packageID == packageID {
			appLoggerAffectedByEnv = true
		}

		registry.setLevelForEntry(entry, level, trace)
	})

	if !appLoggerAffectedByEnv {
		// No environment affected the application logger, let's force INFO to be used for all entries with the same shortName (usually a common project)
		for _, entry := range registry.entriesByShortName[shortName] {
			registry.setLevelForEntry(entry, zapcore.InfoLevel, false)
		}
	}

	// The application logger is guaranteed to be set at this point, at worst it will only be active for >= DPanicLevel
	appLogger := registry.entriesByPackageID[packageID].logPtr

	// Hijack standard Golang `log` and redirects it to our common logger
	zap.RedirectStdLogAt(appLogger, zap.DebugLevel)

	if loggerOptions.switcherServerAutoStart != nil && *loggerOptions.switcherServerAutoStart {
		go func() {
			listenAddr := loggerOptions.switcherServerListenAddr
			appLogger.Info("starting atomic level switcher", zap.String("listen_addr", listenAddr))

			handler := &switcherServerHandler{registry: registry}
			if err := http.ListenAndServe(listenAddr, handler); err != nil {
				appLogger.Warn("failed starting atomic level switcher", zap.Error(err), zap.String("listen_addr", listenAddr))
			}
		}()
	}

	return logger, tracer
}

// NewLogger creates a new logger with sane defaults based on a varity of rules described
// below and automatically registered withing the logging registry.
func NewLogger(opts ...LoggerOption) *zap.Logger {
	options := loggerOptions{}
	for _, opt := range opts {
		opt.apply(&options)
	}

	return newLogger(&options)
}

func MaybeNewLogger(opts ...LoggerOption) (*zap.Logger, error) {
	options := loggerOptions{}
	for _, opt := range opts {
		opt.apply(&options)
	}

	logger, err := maybeNewLogger(&options)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func newLogger(opts *loggerOptions) *zap.Logger {
	logger, err := maybeNewLogger(opts)
	if err != nil {
		panic(fmt.Errorf("unable to create logger (in production? %t): %w", opts.isProductionEnvironment(), err))
	}

	return logger
}

func maybeNewLogger(opts *loggerOptions) (logger *zap.Logger, err error) {
	defer func() {
		if logger != nil && opts.loggerName != "" {
			logger = logger.Named(opts.loggerName)
		}
	}()

	zapOptions := opts.zapOptions

	if opts.isProductionEnvironment() || opts.forceProductionLogger {
		reportAllErrors := opts.reportAllErrors != nil
		serviceName := opts.serviceName

		if reportAllErrors && opts.serviceName != nil {
			zapOptions = append(zapOptions, zapdriver.WrapCore(zapdriver.ReportAllErrors(true), zapdriver.ServiceName(*serviceName)))
		} else if reportAllErrors {
			zapOptions = append(zapOptions, zapdriver.WrapCore(zapdriver.ReportAllErrors(true)))
		} else if opts.serviceName != nil {
			zapOptions = append(zapOptions, zapdriver.WrapCore(zapdriver.ServiceName(*serviceName)))
		}

		config := zapdriver.NewProductionConfig()
		config.Level = *opts.level

		return config.Build(zapOptions...)
	}

	// Development logger
	isTTY := terminal.IsTerminal(int(os.Stderr.Fd()))
	logStdoutWriter := zapcore.Lock(os.Stderr)
	verbosity := 1
	if opts.encoderVerbosity != nil {
		verbosity = *opts.encoderVerbosity
	}

	return zap.New(zapcore.NewCore(NewEncoder(verbosity, isTTY), logStdoutWriter, opts.level), zapOptions...), nil
}

type Tracer interface {
	Enabled() bool
}

type boolTracer struct {
	value *bool
}

func (t boolTracer) Enabled() bool {
	if t.value == nil {
		return false
	}

	return *t.value
}

func ptrBool(value bool) *bool                        { return &value }
func ptrInt(value int) *int                           { return &value }
func ptrString(value string) *string                  { return &value }
func ptrLevel(value zap.AtomicLevel) *zap.AtomicLevel { return &value }
