package log

import (
	"time"

	"github.com/go-cinch/common/log"
)

type Options struct {
	colorful bool
	slow     time.Duration
	level    log.Level
}

func WithColorful(f bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).colorful = f
	}
}

func WithSlow(milli int64) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).slow = time.Duration(milli) * time.Millisecond
	}
}

func WithLevel(level log.Level) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).level = level
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			colorful: false,
			slow:     200 * time.Millisecond,
			level:    log.DebugLevel,
		}
	}
	return options
}
