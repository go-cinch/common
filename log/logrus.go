package log

import (
	"context"
	"fmt"

	"github.com/go-cinch/common/log/caller"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/sirupsen/logrus"
)

var _ log.Logger = (*kratosLog)(nil)

// a Logger implement for override default kratos log
type kratosLog struct {
	log Logger
}

func (l kratosLog) Log(level log.Level, keyvals ...interface{}) (err error) {
	ns := make(Fields)
	var msg string
	for i := 0; i < len(keyvals); i += 2 {
		if keyvals[i] == "" {
			msg = fmt.Sprintf("%v", keyvals[i+1])
			continue
		}
		ns[fmt.Sprintf("%v", keyvals[i])] = keyvals[i+1]
	}
	l.log.
		WithFields(ns).
		Log(kratosLevelToLogLevel(level), msg)
	return
}

var _ Logger = (*logrusLog)(nil)

type logrusLog struct {
	log     *logrus.Entry
	ops     Options
	fields  Fields
	valuers Fields
}

func newLogrusLog(ops *Options) *logrusLog {
	logger := logrus.New()
	logger.SetLevel(loggerToLogrusLogLevel(ops.level))
	logger.SetFormatter(ops.textFormatter)
	if ops.json {
		logger.SetFormatter(ops.jsonFormatter)
	}
	valuers := ops.valuers
	if ops.caller {
		valuers[CallerKey] = caller.Caller(ops.callOptions...)
	}
	entry := logrus.NewEntry(logger)
	l := logrusLog{
		log:     entry,
		ops:     *ops,
		fields:  make(Fields),
		valuers: valuers,
	}
	// override default kratos log
	k := &kratosLog{
		log: &l,
	}
	// clear default message key
	log.DefaultMessageKey = ""
	log.SetLogger(k)
	l.ops.logger = k
	return &l
}

func (l *logrusLog) Options() Options {
	return l.ops
}

func (l *logrusLog) WithFields(fields Fields) Logger {
	ns := copyFields(fields)
	for k, v := range l.fields {
		ns[k] = v
	}
	ll := &logrusLog{
		log:     l.log,
		ops:     l.ops,
		fields:  ns,
		valuers: l.valuers,
	}
	return ll
}

func (l *logrusLog) WithContext(ctx context.Context) Logger {
	ns := copyFields(l.fields)
	ll := &logrusLog{
		log:     l.log.WithContext(ctx),
		ops:     l.ops,
		fields:  ns,
		valuers: l.valuers,
	}
	return ll
}

func (l *logrusLog) Log(level Level, args ...interface{}) {
	ns := l.bindValues()
	// use new entry avoid race
	ll := logrus.
		NewEntry(l.log.Logger).
		WithContext(l.log.Context)
	if len(ns) > 0 {
		ll = ll.WithFields(logrus.Fields(ns))
	}
	ll.Log(loggerToLogrusLogLevel(level), fmt.Sprint(args...))
}

func (l *logrusLog) Logf(level Level, format string, args ...interface{}) {
	ns := l.bindValues()
	ll := logrus.
		NewEntry(l.log.Logger).
		WithContext(l.log.Context)
	if len(ns) > 0 {
		ll = ll.WithFields(logrus.Fields(ns))
	}
	ll.Log(loggerToLogrusLogLevel(level), fmt.Sprintf(format, args...))
}

func (l *logrusLog) bindValues() Fields {
	ns1 := copyFields(l.fields)
	for k, v := range l.valuers {
		var val interface{}
		switch v.(type) {
		case log.Valuer:
			val = v.(log.Valuer)(l.log.Context)
		default:
			val = v
		}
		ns1[k] = val
	}
	ns2 := make(Fields)
	for k, v := range ns1 {
		if str, ok := v.(string); ok && l.ops.skipEmpty && str == "" {
			continue
		}
		ns2[k] = v
	}
	return ns2
}

func loggerToLogrusLogLevel(level Level) logrus.Level {
	switch level {
	case TraceLevel:
		return logrus.TraceLevel
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case FatalLevel:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

func kratosLevelToLogLevel(level log.Level) Level {
	switch level {
	case log.LevelDebug:
		return DebugLevel
	case log.LevelInfo:
		return InfoLevel
	case log.LevelWarn:
		return WarnLevel
	case log.LevelError:
		return ErrorLevel
	case log.LevelFatal:
		return FatalLevel
	default:
		return InfoLevel
	}
}
