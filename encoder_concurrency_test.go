package logging

import (
	"io"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Regression test for a data race in jsonEncoder.clone(): getJSONEncoder()
// hands out a pooled *jsonEncoder that a concurrent clone() call is still
// mutating. Triggered by concurrent Logger.With + log emission on the same
// base logger (the pattern loopagent's Thread.log exercised). Run with -race.
func TestEncoderConcurrentWithAndLog(t *testing.T) {
	core := zapcore.NewCore(NewEncoder(1, false), zapcore.AddSync(io.Discard), zap.DebugLevel)
	base := zap.New(core).Named("base").With(zap.String("thread_id", "t-1"), zap.String("agent_id", "a-1"))

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 2000; j++ {
				derived := base.With(zap.Int("n", n))
				derived.Info("hello", zap.String("j", string(rune(j))))
			}
		}(i)
	}
	wg.Wait()
}
