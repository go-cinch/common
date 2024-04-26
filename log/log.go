package log

import "context"

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// Logger logger interface
type Logger interface {
	Options() Options
	WithFields(fields Fields) Logger
	WithContext(ctx context.Context) Logger
	Log(level Level, v ...interface{})
	Logf(level Level, format string, v ...interface{})
}

type Config struct {
	ops Options
}

func New(options ...func(*Options)) (l Logger) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// l = newKratosLog(ops)
	// if ops.json {
	l = newLogrusLog(ops)
	// }
	return l
}

func copyFields(src Fields) Fields {
	dst := make(Fields, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
