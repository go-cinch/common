package log

import "context"

var DefaultWrapper *Wrapper

func init() {
	DefaultWrapper = NewWrapper()
}

func Trace(args ...interface{}) {
	DefaultWrapper.Trace(args...)
}

func Debug(args ...interface{}) {
	DefaultWrapper.Debug(args...)
}

func Info(args ...interface{}) {
	DefaultWrapper.Info(args...)
}

func Warn(args ...interface{}) {
	DefaultWrapper.Warn(args...)
}

func Error(args ...interface{}) {
	DefaultWrapper.Error(args...)
}

func Fatal(args ...interface{}) {
	DefaultWrapper.Fatal(args...)
}

func WithError(err error) *Wrapper {
	return DefaultWrapper.WithError(err)
}

func WithField(k string, v interface{}) *Wrapper {
	return DefaultWrapper.WithField(k, v)
}

func WithFields(fields Fields) *Wrapper {
	return DefaultWrapper.WithFields(fields)
}

func WithContext(ctx context.Context) *Wrapper {
	return DefaultWrapper.WithContext(ctx)
}
