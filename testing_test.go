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
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTestLogger_Nothing(t *testing.T) {
	logger := NewTestLogger(t)

	require.Equal(t, []string(nil), logger.RecordedLines(t))
}
func TestTestLogger_Single(t *testing.T) {
	logger := NewTestLogger(t)
	testLog(logger.Instance(), "value")

	require.Equal(t, []string{`{"level":"info","msg":"value"}`}, logger.RecordedLines(t))
}

func TestTestLogger_Multi(t *testing.T) {
	logger := NewTestLogger(t)
	testLog(logger.Instance(), "one")
	testLog(logger.Instance(), "two")
	testLog(logger.Instance(), "three")

	require.Equal(t, []string{
		`{"level":"info","msg":"one"}`,
		`{"level":"info","msg":"two"}`,
		`{"level":"info","msg":"three"}`,
	}, logger.RecordedLines(t))
}

func testLog(logger *zap.Logger, message string) {
	logger.Info(message)
}
