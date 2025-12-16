# StreamingFast Logging library
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/logging)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This is the logging library used as part of **[StreamingFast](https://github.com/streamingfast/streamingfast)**.


## Usage

In all library packages (by convention, use the import path):

```go
var zlog *zap.Logger

func init() {
	logging.Register("github.com/path/to/my/package", &zlog)
}
```

In `main` packages:

```go
var zlog *zap.Logger

func setupLogger() {
	logging.Register("main", &zlog)

	logging.Set(logging.MustCreateLogger())
	// Optionally set a different logger here and there,
	// using a regexp matching the registered names:
	//logging.Set(zap.NewNop(), "eosdb")
}
```

In tests (to avoid a race between the `init()` statements)

```go
func init() {
	if os.Getenv("DEBUG") != "" {
		logging.Override(logging.MustCreateLoggerWithLevel("test", zap.NewAtomicLevelAt(zap.DebugLevel)), ""))
	}
}
```

You can switch log levels dynamically, by poking the port 1065 like this:

On listening servers (port 1065, hint: logs!)

* `curl http://localhost:1065/ -XPUT -d '{"level": "debug", "inputs": "github.com/my/package"}'`

## Auto-Reset Feature

The logging library includes an auto-reset feature that automatically resets debug/trace log levels back to INFO after a configurable timeout period. This helps prevent accidentally leaving debug/trace levels enabled, which can cause excessive logging costs.

### Configuration

Auto-reset is enabled by default in production environments with a 30-minute timeout. You can configure it using the following options:

```go
// Enable auto-reset with default 30-minute timeout
logging.InstantiateLoggers(
    logging.WithLogLevelSwitcherServerAutoStart(),
    logging.WithLogLevelAutoResetEnabled(),
)

// Configure custom timeout
logging.InstantiateLoggers(
    logging.WithLogLevelSwitcherServerAutoStart(),
    logging.WithLogLevelAutoResetEnabled(),
    logging.WithLogLevelAutoResetTimeout(15 * time.Minute),
)

// Disable auto-reset (useful in development)
logging.InstantiateLoggers(
    logging.WithLogLevelSwitcherServerAutoStart(),
    logging.WithLogLevelAutoResetDisabled(),
)
```

### HTTP API

The HTTP server accepts the following JSON payload:

```json
{
    "level": "debug",
    "inputs": "github.com/my/package",
    "permanent": false
}
```

- `level`: Log level (trace, debug, info, warn, error)
- `inputs`: Pattern to match logger names (can be regex)
- `permanent`: Optional boolean to disable auto-reset for this specific request

### Examples

```bash
# Set debug level with auto-reset (will reset to INFO after timeout)
curl http://localhost:1065/ -XPUT -d '{"level": "debug", "inputs": "github.com/my/package"}'

# Set debug level permanently (will not auto-reset)
curl http://localhost:1065/ -XPUT -d '{"level": "debug", "inputs": "github.com/my/package", "permanent": true}'

# Set trace level for all loggers with auto-reset
curl http://localhost:1065/ -XPUT -d '{"level": "trace", "inputs": ".*"}'

# Reset to info level (removes from auto-reset tracking)
curl http://localhost:1065/ -XPUT -d '{"level": "info", "inputs": "github.com/my/package"}'
```

### Behavior

- Only DEBUG and TRACE levels are tracked for auto-reset
- INFO, WARN, and ERROR levels are not affected by auto-reset
- When a pattern expires, it's automatically reset to INFO level
- Permanent patterns (with `"permanent": true`) are never auto-reset
- Setting a pattern to INFO/WARN/ERROR removes it from auto-reset tracking
- Pattern updates refresh the timeout timer

### Zapx

We provide the package `zapx` as a way to add some helpers that are called mostly similar as regular zap named field like `zap.String(<name>, <value>)` to make some recurring use cases easier to implement globally.

- `zapx.Secret(<name>, <value>)` print the string obfuscated using a 1/4 ratio for masking the input, see [test cases](./zapx/secrets_test.go) to "see" how it looks like.

## Contributing

**Issues and PR in this repo related strictly to the streamingfast logging library.**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.

## License

[Apache 2.0](LICENSE)
