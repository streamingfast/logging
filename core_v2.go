package logging

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
)

// This v2 version of `core.go` is a work in progress without any backward compatibility
// version. It might not made it to an official version of the library so you can depend
// on it at your own risk.

// We need to **not** be in a `func init` block because Golang first do "free-form" initialization and then the `init` ones
// so if we want to get proper logging when debugging the `logging` library, we need to have that code before any other
// initialization, which must be the case because every other package(s) depends on this one if using `PackageLogger`.
var dbgZlog, _ = debugLoggerForLoggingLibrary()

var globalRegistry = newRegistry("global", dbgZlog)

type instantiateOptions struct {
	consoleOutput                          *string
	defaultLevel                           *zapcore.Level
	logLevelSwitcherServerAutoStart        *bool
	logLevelSwitcherServerListenAddr       string
	logLevelSwitcherServerAutoResetEnabled *bool
	logLevelSwitcherServerAutoResetTimeout *time.Duration
	logToFile                              string
	forceProductionLogger                  bool
	preSpec                                *logLevelSpec
	reportAllErrors                        *bool
	productionLoggerDetector               func() bool

	// Deprecated
	serviceName *string

	// Used internally for testing purposes
	zapOptions []zap.Option
}

func newInstantiateOptions(opts ...InstantiateOption) instantiateOptions {
	options := instantiateOptions{logLevelSwitcherServerListenAddr: "127.0.0.1:1065"}
	for _, opt := range opts {
		opt.apply(&options)
	}

	if options.reportAllErrors == nil {
		WithReportAllErrors().apply(&options)
	}

	if options.logLevelSwitcherServerAutoStart == nil && options.isProductionEnvironment() {
		WithSwitcherServerAutoStart().apply(&options)
	}

	// Enable auto-reset by default with 30 minute timeout
	if options.logLevelSwitcherServerAutoResetEnabled == nil {
		WithLogLevelSwitcherServerAutoResetEnabled().apply(&options)
	}
	if options.logLevelSwitcherServerAutoResetTimeout == nil {
		WithLogLevelSwitcherServerAutoResetTimeout(30 * time.Minute).apply(&options)
	}

	return options
}

func (o instantiateOptions) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("default_level", ptrLevelToString(o.defaultLevel))
	encoder.AddBool("force_production_logger", o.forceProductionLogger)
	encoder.AddString("log_level_switcher_server_auto_start", ptrBoolToString(o.logLevelSwitcherServerAutoStart))
	encoder.AddString("log_level_switcher_server_listen_addr", o.logLevelSwitcherServerListenAddr)
	encoder.AddString("log_level_switcher_server_auto_reset_enabled", ptrBoolToString(o.logLevelSwitcherServerAutoResetEnabled))
	encoder.AddString("log_level_switcher_server_auto_reset_timeout", ptrDurationToString(o.logLevelSwitcherServerAutoResetTimeout))
	encoder.AddString("pre_spec", ptrLogLevelSpecToString(o.preSpec))
	encoder.AddString("report_all_errors", ptrBoolToString(o.reportAllErrors))
	encoder.AddBool("custom_production_logger_detector", o.productionLoggerDetector != nil)

	encoder.AddString("service_name", ptrStringToString(o.serviceName))

	return nil
}

type InstantiateOption interface {
	apply(o *instantiateOptions)
}

type instantiateFuncOption func(o *instantiateOptions)

func (f instantiateFuncOption) apply(o *instantiateOptions) {
	f(o)
}

// WithLogLevelSwitcherServerAutoStart is going to start the HTTP server
// that enables switching log levels dynamically based on a key without
// relying on the built-in production environment detector to determine if
// in production and only then starting the HTTP server.
//
// If not specified, the default behavior is to start the HTTP server
// for dynamic log switching only if the production environment detector
// detected that we are in a production environment.
//
// Once the HTTP server is started, you can use:
//
// curl -XPUT -d '{"level":"debug","inputs":"true"}' http://localhost:1065
//
// Which in this example above, would change all
func WithLogLevelSwitcherServerAutoStart() InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.logLevelSwitcherServerAutoStart = ptrBool(true)
	})
}

// Deprecated: Use `WithLogLevelSwitcherServerAutoStart` instead
func WithSwitcherServerAutoStart() InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.logLevelSwitcherServerAutoStart = ptrBool(true)
	})
}

// WithLogLevelSwitcherServerListeningAddress configures the listening address the HTTP
// server log level switcher listens to if started.
//
// **Note** This does **not** automatically activate the level switcher server,
// you still must used `WithSwitcherServerAutoStart` option or start it manually
// for this option to have any effect.
func WithLogLevelSwitcherServerListeningAddress(addr string) InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.logLevelSwitcherServerListenAddr = addr
	})
}

// Deprecated: Use `WithLogLevelSwitcherServerListeningAddress` instead
func WithSwitcherServerListeningAddress(addr string) InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.logLevelSwitcherServerListenAddr = addr
	})
}

// WithLogLevelSwitcherServerAutoResetEnabled enables the automatic reset of debug/trace log levels
// back to INFO level after a configurable timeout period. This helps prevent
// accidentally leaving debug/trace levels enabled which can cause excessive logging costs.
//
// When enabled, any log level changes to DEBUG or TRACE via the HTTP server will be
// automatically reset to INFO level after the configured timeout (default 30 minutes).
// This can be overridden on a per-request basis using the "permanent" field in the HTTP request.
func WithLogLevelSwitcherServerAutoResetEnabled() InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.logLevelSwitcherServerAutoResetEnabled = ptrBool(true)
	})
}

// WithLogLevelSwitcherServerAutoResetDisabled disables the automatic reset of debug/trace log levels.
// This is useful in development environments where you want debug/trace levels to persist.
func WithLogLevelSwitcherServerAutoResetDisabled() InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.logLevelSwitcherServerAutoResetEnabled = ptrBool(false)
	})
}

// WithLogLevelSwitcherServerAutoResetTimeout configures the timeout duration after which debug/trace
// log levels will be automatically reset to INFO level. The default is 30 minutes.
//
// This setting only takes effect when auto-reset is enabled via WithLogLevelSwitcherServerAutoResetEnabled().
func WithLogLevelSwitcherServerAutoResetTimeout(timeout time.Duration) InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.logLevelSwitcherServerAutoResetTimeout = &timeout
	})
}

func WithReportAllErrors() InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.reportAllErrors = ptrBool(true)
	})
}

// Deprecated: Will be removed in a future version, if your were using that in `ApplicationLogger`,
// use `RootLogger` and set the option there then in a `init` func in your main entry point,
// call `InstantiateLoggers`.
func WithServiceName(name string) InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.serviceName = ptrString(name)
	})
}

// WithProductionLogger enforces the use of the production logger without relying on the built-in
// production environment detector.

// The actual production logger is automatically inferred based on various environmental conditions,
// defaulting to `stackdriver` format which is ultimately just a JSON logger with formatted in such
// that is ingestible by Stackdriver compatible ingestor(s) (todays know as Google Cloud Operations).
func WithProductionLogger() InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.forceProductionLogger = true
	})
}

// WithProductionDetector changes the production environment detector to the one provided
// as argument.
//
// The default production environment detector tries to infer if running inside a container
// through various checks. If you want to use a different production environment detector,
// use this option.
func WithProductionDetector(productionLoggerDetector func() bool) InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.productionLoggerDetector = productionLoggerDetector
	})
}

// WithDefaultLevel is going to set `level` as the default level for all loggers
// instantiated.
func WithDefaultLevel(level zapcore.Level) InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.defaultLevel = ptrLevel(level)
	})
}

// WithDefaultSpec is going to set `level` of the loggers affected by the spec, each entry
// being of the form `<matcher>=<level>` where the `matcher` is the matching input (can be
// a short name directly or a regex matched against short name and package ID second).
//
// This has more precedence over `WithDefaultLevel` which means that it's possible
// to use `WithDefaultLevel(zapcore.InfoLevel)` and then be more specific providing
// `WithDefaultSpec(...)` with some entries.
func WithDefaultSpec(specs ...string) InstantiateOption {
	var logLevelSpec *logLevelSpec
	if len(specs) > 0 {
		logLevelSpec = newLogLevelSpec(envGetFromMap(map[string]string{
			"DLOG": strings.Join(specs, ","),
		}))
	}

	return instantiateFuncOption(func(o *instantiateOptions) {
		if logLevelSpec != nil {
			o.preSpec = logLevelSpec
		}
	})
}

// WithOutputToFile configures the loggers to write to the `logFile` received in the argument
// in **addition** to the console logging that is performed automatically.
//
// The actual format of the log file will a JSON fromat `stackdriver` format which is ultimately
// just a JSON logger with formatted in such that is ingestible by Stackdriver compatible
// ingestor(s) (todays know as Google Cloud Operations).
func WithOutputToFile(logFile string) InstantiateOption {
	if logFile == "" {
		panic(fmt.Errorf("the receive log file value is empty, this is not accepted as a valid option"))
	}

	return instantiateFuncOption(func(o *instantiateOptions) {
		o.logToFile = logFile
	})
}

// WithConsoleToStdout configures the console to log to `stdout` instead of the default
// which is to log to `stderr`.
func WithConsoleToStdout() InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.consoleOutput = ptrString("stdout")
	})
}

// WithConsoleToStderr configures the console to log to `stderr`, which is the default.
func WithConsoleToStderr() InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.consoleOutput = ptrString("stderr")
	})
}

func withZapOption(option zap.Option) InstantiateOption {
	return instantiateFuncOption(func(o *instantiateOptions) {
		o.zapOptions = append(o.zapOptions, option)
	})
}

func (o *instantiateOptions) isProductionEnvironment() bool {
	if o.forceProductionLogger {
		return true
	}

	if o.productionLoggerDetector != nil {
		return o.productionLoggerDetector()
	}

	// Inside Docker runtime, this file is populated
	_, err := os.Stat("/.dockerenv")
	if err == nil {
		return true
	}

	// Inside container runtime the mounts can be inspected to see if we are running inside a container
	if content, err := os.ReadFile("/proc/self/mounts"); err == nil {
		// The containerd runtime mounts the container rootfs under `/var/lib/containerd`
		if bytes.Contains(content, []byte("/var/lib/containerd")) {
			return true
		}

		// The docker runtime mounts the container rootfs under `/var/lib/docker`
		if bytes.Contains(content, []byte("/var/lib/docker")) {
			return true
		}
	}

	// Inside Kubernetes runtime, this env var is set
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	return false
}

// GlobalRegistry returns the global registry where all [PackageLogger] calls register to.
//
// This is primarily intended for cross-module migration scenarios, such as allowing a v2
// module to discover and re-instantiate loggers registered via the v1 [PackageLogger].
func GlobalRegistry() Registry {
	return globalRegistry
}

// PackageLogger creates a new no-op logger (via `zap.NewNop`) and automatically registered it
// withing the logging registry with a tracer that can be be used for conditionally tracing
// code.
//
// The logger actual logger instance will be created when `InstantiateLoggers` is called
// somewhere in the main entry point of your application.
//
// The `shortName` is usually the project name (e.g., "myapp") and the `packageID` is
// usually the full package path (e.g., "github.com/myorg/myapp/mypackage").
//
// Example:
//
//		var zlog, tracer = logging.PackageLogger("myapp", "github.com/myorg/myapp/mypackage")
//
//		func main() {
//			logging.InstantiateLoggers()
//		    ...
//		 }
//
//		func MyFunction() {
//		 	if tracer.Enabled() {
//				// Use debug since zap doesn't have trace level
//				zlog.Debug(...)
//			}
//
//			zlog.Info("some logging", zap.String("key", "value"))
//	  }
//
// You should used this in packages that are not `main` packages
func PackageLogger(shortName string, packageID string, registerOptions ...LoggerOption) (*zap.Logger, Tracer) {
	return packageLogger(globalRegistry, shortName, packageID, registerOptions...)
}

func packageLogger(registry *registry, shortName string, packageID string, registerOptions ...LoggerOption) (*zap.Logger, Tracer) {
	return registry.Register(shortName, packageID, registerOptions...)
}

// InstantiateLoggers correctly instantiate all the loggers of your application at the correct
// level based on various source of information. The source of information that are checked are
//
// - Environment variable (WARN, DEBUG, INFO, ERROR)
// - InstantiateOption passed directly to this method (take precedences over environment variable)
//
// Loggers are created by calling `PackageLogger("<shortName>", "<packageID>")` and are registered
// internally.
//
// Here the set of rules used and the outcome they are giving:
//
//  1. If a production environment is detected (for now, only checking if file /.dockerenv exists)
//     Use a JSON StackDriver compatible format
//
//  2. Otherwise
//     Use a developer friendly colored format
//
// This need some more documentation, it's not documenting the other options that someone can pass
// around.
//
// *Note* The InstantiateLoggers should be called only once per process. That could be enforced
//
//	in the future.
func InstantiateLoggers(opts ...InstantiateOption) {
	instantiateLoggers(globalRegistry, os.Getenv, newInstantiateOptions(opts...))
}

// ApplicationLogger calls `RootLogger` followed by `InstantiateLoggers`. It's a one-liner when
// creating scripts to both create the root logger and instantiate all loggers.
//
// If you require configuring some details of the root logger, make the two calls manually.
//
// Deprecated: This was done with the intent that the "root" logger would be info by default and all
// others wouldn't. Trying to be smart just causes confusion at the API level. Always use [PackageLogger]
// and then call [InstantiateLoggers] and use [WithDefaultSpec] to control finely what you want in term of
// spec like `WithDefaultSpec("logger1=info,logger2=debug,.*=warn")` (regex supported).
func ApplicationLogger(shortName string, packageID string, opts ...InstantiateOption) (*zap.Logger, Tracer) {
	return applicationLogger(globalRegistry, os.Getenv, shortName, packageID, opts...)
}

func applicationLogger(registry *registry, envGet func(string) string, shortName string, packageID string, opts ...InstantiateOption) (*zap.Logger, Tracer) {
	options := newInstantiateOptions(opts...)
	logger, tracer := rootLogger(registry, shortName, packageID)

	instantiateLoggers(registry, envGet, options)

	return logger, tracer
}

// RootLogger should be used to get a logger for a top-level binary application which will
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
// Deprecated: This was done with the intent that the "root" logger would be info by default and all
// others wouldn't. Trying to be smart just causes confusion at the API level. Always use [PackageLogger]
// and when calling [InstantiateLoggers], use [WithDefaultSpec] to control finely what you want in term of
// spec like `WithDefaultSpec("logger1=info,logger2=debug,.*=warn")` (regex supported).
func RootLogger(shortName string, packageID string, opts ...LoggerOption) (*zap.Logger, Tracer) {
	return rootLogger(globalRegistry, shortName, packageID, opts...)
}

func rootLogger(registry *registry, shortName string, packageID string, opts ...LoggerOption) (*zap.Logger, Tracer) {
	return registry.Register(shortName, packageID, append(opts, loggerRoot())...)
}

func instantiateLoggers(registry *registry, envGet func(string) string, options instantiateOptions) {
	dbgZlog.Info("instantiate loggers invoked", zap.Object("options", options))

	// We override the factory function so that we use "our" options which are those passed by the
	// developer.
	registry.factory = func(name string, level zap.AtomicLevel) *zap.Logger {
		return newLogger(registry.dbgLogger, name, level, &options)
	}

	dbgZlog.Info("creating all loggers")
	registry.forAllEntries(func(entry *registryEntry) {
		registry.createLoggerForEntry(entry)
	})

	rootLoggerAffectedByUser := false

	if options.defaultLevel != nil {
		dbgZlog.Info("override level from default level option")
		registry.forAllEntries(func(entry *registryEntry) {
			registry.setLevelForEntry(entry, *options.defaultLevel, false)
		})

		// Root logger is always affected by this, since the default level affects all loggers
		rootLoggerAffectedByUser = true
	}

	// We first override the level based on pre spec passed by the developer on the InstantiateLoggers, if set
	if options.preSpec != nil {
		dbgZlog.Info("override level from pre spec option", zap.Bool("has_root_logger", registry.rootEntry != nil))

		registry.forAllEntriesMatchingSpec(options.preSpec, func(entry *registryEntry, level zapcore.Level, trace bool) {
			if registry.rootEntry != nil && entry.packageID == registry.rootEntry.packageID {
				dbgZlog.Info("root logger affected by env", zap.Stringer("root", registry.rootEntry))
				rootLoggerAffectedByUser = true
			}

			dbgZlog.Debug("setting logger entry matching from pre spec with level logger", zap.Stringer("to_level", level), zap.Stringer("entry", entry))
			registry.setLevelForEntry(entry, level, trace)
		})
	}

	// We then override the level based on the spec extracted from the environment
	dbgZlog.Info("override level from env specification", zap.Bool("has_root_logger", registry.rootEntry != nil))

	logLevelSpec := newLogLevelSpec(envGet)
	registry.forAllEntriesMatchingSpec(logLevelSpec, func(entry *registryEntry, level zapcore.Level, trace bool) {
		if registry.rootEntry != nil && entry.packageID == registry.rootEntry.packageID {
			dbgZlog.Info("root logger affected by env", zap.Stringer("root", registry.rootEntry))
			rootLoggerAffectedByUser = true
		}

		dbgZlog.Debug("setting logger entry matching from env with level logger", zap.Stringer("to_level", level), zap.Stringer("entry", entry))
		registry.setLevelForEntry(entry, level, trace)
	})

	rootLogger := zap.NewNop()

	if registry.rootEntry != nil {
		rootLogger = registry.rootEntry.logPtr

		if !rootLoggerAffectedByUser {
			// No environment affected the root logger, let's force INFO to be used for all entries with the same shortName (usually a common project)
			for _, entry := range registry.entriesByShortName[registry.rootEntry.shortName] {
				dbgZlog.Debug("setting logger by short name with info logger because the root logger has not been affected by any env",
					zap.Stringer("to_level", zap.InfoLevel),
					zap.Stringer("entry", entry),
				)

				registry.setLevelForEntry(entry, zapcore.InfoLevel, false)
			}
		}
	}

	// Hijack standard Golang `log` and redirects it to our common logger
	zap.RedirectStdLogAt(rootLogger, zap.DebugLevel)

	if options.logLevelSwitcherServerAutoStart != nil && *options.logLevelSwitcherServerAutoStart {
		go func() {
			listenAddr := options.logLevelSwitcherServerListenAddr
			dbgZlog.Info("starting atomic level switcher", zap.String("listen_addr", listenAddr))

			// Initialize pattern tracker if auto-reset is enabled
			var patternTracker *patternTracker
			if options.logLevelSwitcherServerAutoResetEnabled != nil && *options.logLevelSwitcherServerAutoResetEnabled {
				timeout := 30 * time.Minute // default timeout
				if options.logLevelSwitcherServerAutoResetTimeout != nil {
					timeout = *options.logLevelSwitcherServerAutoResetTimeout
				}

				patternTracker = newPatternTracker(registry, timeout, dbgZlog.Named("pattern_tracker"))
				patternTracker.start()

				dbgZlog.Info("started pattern tracker for auto-reset functionality",
					zap.Duration("timeout", timeout))
			}

			handler := &switcherServerHandler{
				registry:       registry,
				patternTracker: patternTracker,
			}

			if err := http.ListenAndServe(listenAddr, handler); err != nil {
				dbgZlog.Warn("failed starting atomic level switcher", zap.Error(err), zap.String("listen_addr", listenAddr))
			}

			// Clean shutdown of pattern tracker
			if patternTracker != nil {
				patternTracker.stop()
			}
		}()
	}

	// Any PackageLogger registered after InstantiateLoggers was called should be
	// immediately instantiated and have the same level rules applied to it.
	capturedEnvSpec := logLevelSpec
	capturedRootAffected := rootLoggerAffectedByUser
	registry.onNewEntry = func(entry *registryEntry) {
		registry.createLoggerForEntry(entry)

		if options.defaultLevel != nil {
			registry.setLevelForEntry(entry, *options.defaultLevel, false)
		}

		if options.preSpec != nil {
			registry.applyLevelSpecToEntry(entry, options.preSpec)
		}

		registry.applyLevelSpecToEntry(entry, capturedEnvSpec)

		if registry.rootEntry != nil && !capturedRootAffected && entry.shortName == registry.rootEntry.shortName {
			registry.setLevelForEntry(entry, zapcore.InfoLevel, false)
		}
	}

	if registry.dbgLogger.Core().Enabled(zapcore.InfoLevel) {
		registry.dumpRegistryToLogger()
	}
}

func newLogger(dbgLogger *zap.Logger, name string, level zap.AtomicLevel, opts *instantiateOptions) *zap.Logger {
	logger, err := maybeNewLogger(dbgLogger, name, level, opts)
	if err != nil {
		panic(fmt.Errorf("unable to create logger (in production? %t): %w", opts.isProductionEnvironment(), err))
	}

	return logger
}

func maybeNewLogger(dbgLogger *zap.Logger, name string, level zap.AtomicLevel, opts *instantiateOptions) (logger *zap.Logger, err error) {
	if name != "" {
		dbgLogger = dbgLogger.With(zap.String("for", name))
	}

	defer func() {
		if logger != nil && name != "" {
			logger = logger.Named(name)
		}
	}()

	consoleOutput := os.Stderr
	if opts.consoleOutput != nil {
		switch *opts.consoleOutput {
		case "stderr":
			consoleOutput = os.Stderr
		case "stdout":
			consoleOutput = os.Stdout
		}
	}

	zapOptions := opts.zapOptions
	isTTY := terminal.IsTerminal(int(consoleOutput.Fd()))
	logConsoleWriter := zapcore.Lock(consoleOutput)

	var fileSyncer zapcore.WriteSyncer
	if opts.logToFile != "" {
		dbgLogger.Debug("creating file syncer", zap.String("log_file", opts.logToFile))

		var err error
		fileSyncer, err = createLogFileWriter(opts.logToFile)
		if err != nil {
			return nil, fmt.Errorf("create file syncer: %w", err)
		}
	}

	var consoleCore zapcore.Core
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

		consoleCore = zapcore.NewCore(zapcore.NewJSONEncoder(zapdriver.NewProductionEncoderConfig()), logConsoleWriter, level)
	} else {
		consoleCore = zapcore.NewCore(NewEncoder(1, isTTY), logConsoleWriter, level)
	}

	if fileSyncer == nil {
		dbgLogger.Debug("returning only console syncer into a standard core, as there is no file syncer defined")
		return zap.New(consoleCore, zapOptions...), nil
	}

	// FIXME: The log to file is always performed in JSON, we should enable some configuration for it to tweak
	// its output format, but it's not clear how this would look like, probably would come with some "general"
	// formatting option that would enabled changing for example the console output format itself.
	dbgLogger.Debug("merging console and file syncer into a tee core")
	fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), fileSyncer, level)
	teeCore := zapcore.NewTee(consoleCore, fileCore)

	return zap.New(teeCore, zapOptions...), nil
}

func createLogFileWriter(logFile string) (zapcore.WriteSyncer, error) {
	err := os.MkdirAll(filepath.Dir(logFile), 0755)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("make directories for log file %q: %w", logFile, err)
	}

	writer, _, err := zap.Open(logFile)
	if err != nil {
		return nil, fmt.Errorf("open log file %q: %w", logFile, err)
	}

	return writer, err
}

type Tracer interface {
	Enabled() bool
}

type boolTracer struct {
	value *bool
}

//go:inline
func (t boolTracer) Enabled() bool {
	if t.value == nil {
		return false
	}

	return *t.value
}

func ptrBool(value bool) *bool                    { return &value }
func ptrString(value string) *string              { return &value }
func ptrLevel(value zapcore.Level) *zapcore.Level { return &value }

func ptrBoolToString(value *bool) string {
	switch {
	case value == nil:
		return "<nil>"
	case *value:
		return "true"
	default:
		return "false"
	}
}

func ptrStringToString(value *string) string {
	if value == nil {
		return "<nil>"
	}

	return *value
}

func ptrLevelToString(value *zapcore.Level) string {
	if value == nil {
		return "<nil>"
	}

	return (*value).String()
}

func ptrLogLevelSpecToString(value *logLevelSpec) string {
	if value == nil {
		return "<nil>"
	}

	return (*value).String()
}

func ptrDurationToString(value *time.Duration) string {
	if value == nil {
		return "<nil>"
	}

	return (*value).String()
}
