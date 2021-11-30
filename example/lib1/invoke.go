package lib1

import (
	"time"
)

func Invoke() {
	zlog.Info("Info level from 'lib'")

	time.Sleep(2 * time.Second)
	zlog.Debug("Debug level from 'lib'")

	if tracer.Enabled() {
		zlog.Debug("Trace level from 'lib'")
	}
}
