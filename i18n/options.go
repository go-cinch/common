package i18n

import (
	"embed"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Options struct {
	format   map[string]i18n.UnmarshalFunc
	language language.Tag
	files    []string
	fs       embed.FS
}

func WithFormat(format string, f i18n.UnmarshalFunc) func(*Options) {
	return func(options *Options) {
		if format != "" && f != nil {
			getOptionsOrSetDefault(options).format[format] = f
		}
	}
}

func WithLanguage(lang language.Tag) func(*Options) {
	return func(options *Options) {
		if lang.String() != "und" {
			getOptionsOrSetDefault(options).language = lang
		}
	}
}

func WithFile(f string) func(*Options) {
	return func(options *Options) {
		if f != "" {
			getOptionsOrSetDefault(options).files = append(getOptionsOrSetDefault(options).files, f)
		}
	}
}

func WithFs(fs embed.FS) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).fs = fs
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			format:   make(map[string]i18n.UnmarshalFunc),
			language: language.English,
			files:    []string{},
		}
	}
	return options
}
