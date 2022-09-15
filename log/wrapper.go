package log

import (
	"context"
	"os"
)

type Wrapper struct {
	log Logger
}

func NewWrapper(options ...func(*Options)) *Wrapper {
	return &Wrapper{
		log: New(options...),
	}
}

func (w *Wrapper) Trace(args ...interface{}) {
	w.print(TraceLevel, args...)
}

func (w *Wrapper) Debug(args ...interface{}) {
	w.print(DebugLevel, args...)
}

func (w *Wrapper) Info(args ...interface{}) {
	w.print(InfoLevel, args...)
}

func (w *Wrapper) Warn(args ...interface{}) {
	w.print(WarnLevel, args...)
}

func (w *Wrapper) Error(args ...interface{}) {
	w.print(ErrorLevel, args...)
}

func (w *Wrapper) Fatal(args ...interface{}) {
	w.print(FatalLevel, args...)
	os.Exit(1)
}

func (w *Wrapper) WithError(err error) *Wrapper {
	return &Wrapper{
		log: w.log.WithFields(Fields{"err": err}),
	}
}

func (w *Wrapper) WithField(k string, v interface{}) *Wrapper {
	return w.WithFields(Fields{
		k: v,
	})
}

func (w *Wrapper) WithFields(fields Fields) *Wrapper {
	return &Wrapper{
		log: w.log.WithFields(fields),
	}
}

func (w *Wrapper) WithContext(ctx context.Context) *Wrapper {
	return &Wrapper{
		log: w.log.WithContext(ctx),
	}
}

func (w *Wrapper) Options() Options {
	return w.log.Options()
}

func (w *Wrapper) print(level Level, args ...interface{}) {
	if !w.log.Options().level.Enabled(level) {
		return
	}
	if len(args) > 1 {
		if format, ok := args[0].(string); ok {
			w.log.Logf(level, format, args[1:]...)
			return
		}
	}
	w.log.Log(level, args...)
}
