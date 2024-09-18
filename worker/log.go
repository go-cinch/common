package worker

import (
	"github.com/go-cinch/common/log"
	"github.com/hibiken/asynq"
)

var _ asynq.Logger = (*myLogger)(nil)

type myLogger struct {
}

func (myLogger) Debug(args ...interface{}) {
	log.Debug(args...)
}

func (myLogger) Info(args ...interface{}) {
	log.Info(args...)
}

func (myLogger) Warn(args ...interface{}) {
	log.Warn(args...)
}

func (myLogger) Error(args ...interface{}) {
	log.Error(args...)
}

func (myLogger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func levelToAsynq(l log.Level) asynq.LogLevel {
	switch l {
	case log.FatalLevel:
		return asynq.FatalLevel
	case log.ErrorLevel:
		return asynq.ErrorLevel
	case log.WarnLevel:
		return asynq.WarnLevel
	case log.InfoLevel:
		return asynq.InfoLevel
	case log.DebugLevel, log.TraceLevel:
		return asynq.DebugLevel
	default:
		return asynq.InfoLevel
	}
}
