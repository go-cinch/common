package i18n

import "golang.org/x/text/language"

type Options struct {
	format   string
	language language.Tag
}

func WithFormat(format string) func(*Options) {
	return func(options *Options) {
		if format != "" {
			getOptionsOrSetDefault(options).format = format
		}
	}
}

func WithLanguage(lang language.Tag) func(*Options) {
	return func(options *Options) {
		if lang.String() != "" {
			getOptionsOrSetDefault(options).language = lang
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			format:   "yml",
			language: language.English,
		}
	}
	return options
}
