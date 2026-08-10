# Change log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Fixed

* Fixed the reflection buffer used by `zap.Any`/`zap.Reflect` fields never being returned to the buffer pool by the developer-friendly `Encoder`, which made every such log line allocate a fresh 1 KiB buffer (measured 1451 B/op down to 392 B/op, 13 allocs down to 11).

### Removed

* Removed the unused `_jsonPool`/`_loggerPool` encoder pools. Nothing ever returned an encoder to them, so they only ever allocated; `jsonEncoder.clone` now allocates directly, which also removes a shared mutable state hazard. No API change.

## v1.2.2

### Added

* `Registry.GetTracerByPackageID(packageID string) (*bool, bool)` exposes the raw trace-enabled pointer for a registered logger, allowing external bridges (e.g. a v2 compatibility layer) to mirror trace state into v1 tracers.
* `logging.LoggerFromContext(ctx, fallback)` as the canonical replacement for `logging.Logger(ctx, fallback)`.

### Deprecated

* `logging.Logger(ctx, fallback)` — use `logging.LoggerFromContext` instead, or the one-liner level helpers `logging.Info`, `logging.Debug`, etc.

## v1.2.1

### Added

* `zapx.TruncatedStringer`, an lazy version of `zapx.TruncatedString`.

## v1.2.0

### Added

* `GlobalRegistry() Registry` exposes the global registry, allowing a future v2 module to discover and re-instantiate all loggers registered via v1 `PackageLogger`.
* `Registry.All(fn func(packageID, shortName string))` iterates over all registered loggers in a registry, enabling cross-module migration scenarios.
* Loggers registered via `PackageLogger` after `InstantiateLoggers` has already been called are now immediately instantiated and have the same level rules (default level, pre-spec, env spec) applied to them.

### Fixed

* `SetLevelFor` now re-applies the current environment variable overrides (DLOG, DEBUG, TRACE, …) after setting the level, so the provided level acts as a baseline that env vars can override. This fixes the common pattern of calling `SetLevelFor(".*", zap.WarnLevel, false)` in a Cobra `PreRun` hook while still allowing `DLOG`/`DEBUG`/`TRACE` to take effect.

### Changed

* The default text `encoder` use to encode log entries now emits the level when coloring is disabled.
* **Deprecated** `logging.IsTraceEnabled`, define your logger and `Tracer` directly with `var zlog, tracer = logging.PackageLogger(<shortName>, "...")` instead of separately, `tracer.Enabled()` can then be used to determine if tracing should be enabled (can be enable dynamically).
* **Deprecated** `logging.TestingOverride`, use `logging.InstantiateLoggers` directly.
* **Deprecated** `logging.Overidde`, use `logging.InstantiateLoggers` directly and use the `logging.WithDefaultSpec` to configure the various loggers.` instead.
* **Deprecated** `logging.Register`, use `var zlog, _ = logging.PackageLogger(<shortName>, "...")` instead.
* **Deprecated** `logging.RegisterOnUpdate`, use `logging.LoggerOnUpdate` instead (will probably be removed actually entirely since it's not needed anymore).
* **Deprecated** `logging.WithServiceName`, no replacement yet (will be `logging.LoggerServiceName` in a future release, if unspecified `shortName` will be used).
* **Deprecated** `logging.RootLogger`, use `logging.PackageLogger` and control levels via `logging.WithDefaultSpec(...)` in `InstantiateLoggers` instead (e.g. `WithDefaultSpec("logger1=info,logger2=debug,.*=warn")`). The root logger concept tried to be smart about defaulting to `info` which caused API confusion.

### Removed

* **BREAKING CHANGE** Removed `logging.Handler`, it has been moved to `dtracing` package to limit transitive depdendencies on this project, you should use `dtracing.NewAddTraceIDAwareLoggerMiddleware` instead.

## 2020-03-21

### Changed

* License changed to Apache 2.0