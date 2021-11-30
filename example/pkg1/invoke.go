package pkg1

import (
	"time"
)

func Invoke() {
	zlog.Info("Info level from 'pkg1'")

	time.Sleep(2 * time.Second)
	zlog.Debug("Debug level from 'pkg1'")

	if tracer.Enabled() {
		zlog.Debug("Trace level from 'pkg1'")
	}
}
