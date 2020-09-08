// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

type LoggerExtender func(*zap.Logger) *zap.Logger

var registry = map[string]**zap.Logger{}
var defaultLogger = zap.NewNop()

func Register(name string, zlogPtr **zap.Logger) {
	if _, found := registry[name]; found {
		panic(fmt.Sprintf("name already registered: %s", name))
	}

	registry[name] = zlogPtr
	*zlogPtr = defaultLogger
}

func Set(logger *zap.Logger, regexps ...string) {
	for name, zlogPtr := range registry {
		if len(regexps) == 0 {
			*zlogPtr = logger
		} else {
			for _, re := range regexps {
				if regexp.MustCompile(re).MatchString(name) {
					*zlogPtr = logger
				}
			}
		}
	}
}

// Extend is different than `Set` by being able to re-configure the existing logger set for
// all registered logger in the registry. This is useful for example to add a field to the
// currently set logger:
//
// ```
// logger.Extend(func (current *zap.Logger) { return current.With("name", "value") }, "github.com/dfuse-io/app.*")
// ```
func Extend(extender LoggerExtender, regexps ...string) {
	for name, zlogPtr := range registry {
		if *zlogPtr == nil {
			continue
		}

		if len(regexps) == 0 {
			*zlogPtr = extender(*zlogPtr)
		} else {
			for _, re := range regexps {
				if regexp.MustCompile(re).MatchString(name) {
					*zlogPtr = extender(*zlogPtr)
				}
			}
		}
	}
}

// Override sets the given logger on previously registered and next
// registrations.  Useful in tests.
func Override(logger *zap.Logger) {
	defaultLogger = logger
	Set(logger)
}

// TestingOverride calls `Override` (or `Set`, see below) with a development
// logger setup correctly with the right level based on some environment variables.
//
// By default, override using a `zap.NewDevelopment` logger (`info`), if
// environment variable `DEBUG` is set to anything or environment variable `TRACE`
// is set to `true`, logger is set in `debug` level.
//
// If `DEBUG` is set to something else than `true` and/or if `TRACE` is set
// to something else than
func TestingOverride() {
	debug := os.Getenv("DEBUG")
	trace := os.Getenv("TRACE")
	if debug == "" && trace == "" {
		return
	}

	logger, _ := zap.NewDevelopment()

	regex := ""
	if debug != "true" {
		regex = debug
	}

	if regex == "" && trace != "true" {
		regex = trace
	}

	if regex == "" {
		Override(logger)
	} else {
		for _, regexPart := range strings.Split(regex, ",") {
			regexPart = strings.TrimSpace(regexPart)
			if regexPart != "" {
				Set(logger, regexPart)
			}
		}
	}
}
