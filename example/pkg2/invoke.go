package pkg2

import (
	"time"
)

func Invoke() {
	zlog.Info("Info level from 'pkg2'")

	time.Sleep(2 * time.Second)
	zlog.Debug("Debug level from 'pkg2'")

	if tracer.Enabled() {
		zlog.Debug("Trace level from 'pkg2'")
	}
}
