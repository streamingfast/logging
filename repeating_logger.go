package logging

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Usage:
// repeatingLogger := logging.NewRepeatingLogger(zlog, repeatingLoggerConfig{RepeatLevel: zap.DebugLevel, RepeatEach: 30 * time.Second, ResetToFirstEach: 3 * time.Minute})

type RepeatingLoggerConfig struct {
	FirstLevel       zapcore.Level
	RepeatLevel      zapcore.Level
	RepeatEach       time.Duration
	ResetToFirstEach time.Duration
}

type RepeatingLogger struct {
	startTime        time.Time
	lastRepeatedTime time.Time
	firstPassed      bool

	logger *zap.Logger
	config RepeatingLoggerConfig
}

func NewRepeatingLogger(logger *zap.Logger, config RepeatingLoggerConfig) *RepeatingLogger {
	return &RepeatingLogger{
		startTime:        time.Now(),
		lastRepeatedTime: time.Now(),

		logger: logger,
		config: config,
	}
}

func (l *RepeatingLogger) Log(msg string, fields ...zapcore.Field) {
	if l.firstPassed && l.config.ResetToFirstEach != 0 && time.Since(l.startTime) > l.config.ResetToFirstEach {
		l.startTime = time.Now()
		l.lastRepeatedTime = time.Now()
		l.firstPassed = false
	}

	level := l.config.RepeatLevel
	if !l.firstPassed {
		level = l.config.FirstLevel
	}

	shouldRepeat := true
	if l.config.RepeatEach != 0 {
		if time.Since(l.lastRepeatedTime) > l.config.RepeatEach {
			l.lastRepeatedTime = time.Now()
		} else {
			shouldRepeat = false
		}
	}

	if !l.firstPassed || shouldRepeat {
		l.logger.Check(level, msg).Write(fields...)
	}

	l.firstPassed = true
}
