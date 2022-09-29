package log

import "github.com/go-kratos/kratos/v2/log"

type Options struct {
	level            Level
	logger           log.Logger
	loggerMessageKey string
	caller           bool
	callerSkip       []string
	callerSource     bool
	callerPrefix     string
	callerLevel      int
	callerVersion    bool
}

func (o Options) Level() Level {
	return o.level
}

func (o Options) Logger() log.Logger {
	return o.logger
}

func WithLevel(level Level) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).level = level
	}
}

func WithLogger(logger log.Logger) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).logger = logger
	}
}

func WithLoggerMessageKey(key string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).loggerMessageKey = key
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			level:            DebugLevel,
			logger:           log.DefaultLogger,
			loggerMessageKey: log.DefaultMessageKey,
			caller:           true,
			callerSkip:       []string{"gorm.io", "go-kratos", "golang.org/x/sync", "go-cinch/common"},
			callerSource:     false,
			callerLevel:      2,
			callerVersion:    true,
		}
	}
	return options
}
