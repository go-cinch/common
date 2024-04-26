package log

import (
	"github.com/go-cinch/common/log/caller"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/sirupsen/logrus"
)

type Options struct {
	level         Level
	logger        log.Logger
	caller        bool
	callOptions   []func(*caller.Options)
	text          bool
	textFormatter *logrus.TextFormatter
	json          bool
	jsonFormatter *logrus.JSONFormatter
	skipEmpty     bool
	valuers       Fields
}

func (o Options) Level() Level {
	return o.level
}

func (o Options) Logger() log.Logger {
	return o.logger
}

// WithLevel change log level, default info
func WithLevel(level Level) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).level = level
	}
}

// WithCaller enable caller
func WithCaller(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).caller = flag
	}
}

// WithCallerOptions add call options
func WithCallerOptions(ops ...func(*caller.Options)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).callOptions = append(getOptionsOrSetDefault(options).callOptions, ops...)
	}
}

// WithValuer print k for each line
func WithValuer(k string, v interface{}) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).valuers[k] = v
	}
}

// WithText enable text format
func WithText(flag bool) func(*Options) {
	return func(options *Options) {
		ops := getOptionsOrSetDefault(options)
		ops.text = flag
		ops.json = !flag
	}
}

// WithJSON enable JSON format
func WithJSON(flag bool) func(*Options) {
	return func(options *Options) {
		ops := getOptionsOrSetDefault(options)
		ops.json = flag
		ops.text = !flag
	}
}

// WithSkipEmpty when key is empty string, not print key
func WithSkipEmpty(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).skipEmpty = flag
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			level:  InfoLevel,
			caller: true,
			text:   true,
			textFormatter: &logrus.TextFormatter{
				FullTimestamp:    true,
				DisableColors:    true,
				DisableQuote:     true,
				DisableSorting:   false,
				QuoteEmptyFields: false,
			},
			json: false,
			jsonFormatter: &logrus.JSONFormatter{
				PrettyPrint: false,
			},
			skipEmpty: true,
			valuers:   make(Fields),
		}
	}
	return options
}
