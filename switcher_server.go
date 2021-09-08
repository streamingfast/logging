package logging

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logChangeReq struct {
	Inputs string `json:"inputs"`
	Level  string `json:"level"`
}

type switcherServerHandler struct {
	registry *registry
}

func (h *switcherServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	in := logChangeReq{}
	if err := decoder.Decode(&in); err != nil {
		http.Error(w, fmt.Sprintf("cannot unmarshal request: %s", err), 400)
		return
	}

	if in.Inputs == "" {
		http.Error(w, "inputs not defined, should be comma-separated list of words or a regular expressions", 400)
		return
	}

	switch strings.ToLower(in.Level) {
	case "warn", "warning":
		h.changeLoggersLevel(in.Inputs, zap.WarnLevel, disableTracing)
	case "info":
		h.changeLoggersLevel(in.Inputs, zap.InfoLevel, disableTracing)
	case "debug":
		h.changeLoggersLevel(in.Inputs, zap.DebugLevel, disableTracing)
	case "trace":
		h.changeLoggersLevel(in.Inputs, zap.DebugLevel, enableTracing)
	default:
		http.Error(w, fmt.Sprintf("invalid level value %q", in.Level), 400)
		return
	}

	w.Write([]byte("ok"))
}

func (h *switcherServerHandler) changeLoggersLevel(inputs string, level zapcore.Level, tracing tracingType) {
	extender := overrideLoggerLevel(level)

	for _, input := range strings.Split(inputs, ",") {
		if entries, found := h.registry.entriesByShortName[input]; found {
			for _, entry := range entries {
				if *entry.logPtr == nil {
					continue
				}

				setLogger(entry, extender(*entry.logPtr), tracing)
			}
		} else {
			extend(overrideLoggerLevel(level), tracing, input)
		}
	}
}

func overrideLoggerLevel(level zapcore.Level) LoggerExtender {
	return func(current *zap.Logger) *zap.Logger {
		return current.WithOptions(WithLevel(level))
	}
}
