Logging library used throughout the `dfuse` platform
----------------------------------------------------


### Usage

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


### Enable debug-level


On listening servers (port 1065, hint: logs!)

* `curl http://localhost:1065/ -XPUT -d '{"level": "debug"}'`
