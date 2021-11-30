package logging

import (
	"os"

	"go.uber.org/zap/zapcore"
)

func init() {
	spec := newLogLevelSpecFromMap(map[string]string{
		"TRACE":   os.Getenv("LOGGING_TRACE"),
		"DEBUG":   os.Getenv("LOGGING_DEBUG"),
		"INFO":    os.Getenv("LOGGING_INFO"),
		"WARNING": os.Getenv("LOGGING_WARNING"),
		"WARN":    os.Getenv("LOGGING_WARN"),
		"ERROR":   os.Getenv("LOGGING_ERROR"),
		"DLOG":    os.Getenv("LOGGING_DLOG"),
	})

	dbgRegistry.forAllEntriesMatchingSpec(spec, func(entry *registryEntry, level zapcore.Level, trace bool) {
		dbgRegistry.setLoggerForEntry(entry, level, trace)
	})
}
