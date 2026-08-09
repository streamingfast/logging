package logging

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// countingSyncer is a `zapcore.WriteSyncer` that only records how many lines it
// received, so the concurrency test can assert nothing was dropped without
// serializing the encoder behind a shared buffer's growth.
type countingSyncer struct {
	mu    sync.Mutex
	lines int
	bytes int
}

func (s *countingSyncer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lines++
	s.bytes += len(p)

	return len(p), nil
}

func (s *countingSyncer) Sync() error { return nil }

// TestEncoderConcurrentWithAndLog exercises `Encoder.Clone` and
// `Encoder.EncodeEntry` from several goroutines sharing one base logger, which is
// how `zapcore.Core.With` and log emission interleave in practice.
//
// It is a guard against shared mutable state creeping back into the encoder (a
// pooled or otherwise reused `jsonEncoder`), not a reproduction of a known race:
// it passes on the pooled implementation too, because that pool was never fed.
// Meaningful under `-race`.
func TestEncoderConcurrentWithAndLog(t *testing.T) {
	const goroutines = 8
	const iterations = 2000

	syncer := &countingSyncer{}
	core := zapcore.NewCore(NewEncoder(1, false), syncer, zap.DebugLevel)
	base := zap.New(core).Named("base").With(zap.String("thread_id", "t-1"), zap.String("agent_id", "a-1"))

	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			for j := range iterations {
				// `With` clones the encoder, the log call clones it again while encoding.
				derived := base.With(zap.Int("n", n))

				// `zap.Any` on a non-primitive goes through the encoder's lazily allocated
				// reflection buffer, covering that path too.
				derived.Info("hello", zap.String("j", strconv.Itoa(j)), zap.Any("payload", map[string]int{"j": j}))
			}
		}(i)
	}
	wg.Wait()

	require.Equal(t, goroutines*iterations, syncer.lines, "every log entry should reach the syncer exactly once")
	assert.Positive(t, syncer.bytes, "encoded entries should not be empty")
}
