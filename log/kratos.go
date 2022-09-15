package log

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
)

var _ Logger = (*kratosLog)(nil)

type kratosLog struct {
	log    *log.Helper
	ops    Options
	fields Fields
}

func newKratosLog(ops *Options) *kratosLog {
	helper := log.NewHelper(ops.logger)
	l := kratosLog{
		log:    helper,
		ops:    *ops,
		fields: make(Fields),
	}
	return &l
}

func (l *kratosLog) Options() Options {
	return l.ops
}

func (l *kratosLog) WithFields(fields Fields) Logger {
	ns := copyFields(fields)
	for k, v := range l.fields {
		ns[k] = v
	}
	ll := &kratosLog{
		log:    l.log,
		ops:    l.ops,
		fields: ns,
	}
	return ll
}

func (l *kratosLog) WithContext(ctx context.Context) Logger {
	ns := copyFields(l.fields)
	ll := &kratosLog{
		log:    l.log.WithContext(ctx),
		ops:    l.ops,
		fields: ns,
	}
	return ll
}

func (l *kratosLog) Log(level Level, args ...interface{}) {
	a := make([]interface{}, 0)
	a = append(a, l.ops.loggerMessageKey, fmt.Sprint(args...))
	for k, v := range l.fields {
		a = append(a, k, v)
	}
	l.log.Log(loggerToKratosLogLevel(level), a...)
}

func (l *kratosLog) Logf(level Level, format string, args ...interface{}) {
	a := make([]interface{}, 0)
	a = append(a, l.ops.loggerMessageKey, fmt.Sprintf(format, args...))
	for k, v := range l.fields {
		a = append(a, k, v)
	}
	l.log.Log(loggerToKratosLogLevel(level), a...)
}

func loggerToKratosLogLevel(level Level) log.Level {
	switch level {
	case TraceLevel, DebugLevel:
		return log.LevelDebug
	case InfoLevel:
		return log.LevelInfo
	case WarnLevel:
		return log.LevelWarn
	case ErrorLevel:
		return log.LevelError
	case FatalLevel:
		return log.LevelFatal
	default:
		return log.LevelInfo
	}
}
