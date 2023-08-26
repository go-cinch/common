package env

type Options struct {
	prefix    string
	separator string
	loaded    func(string, interface{})
}

func WithPrefix(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).prefix = s
	}
}

func WithSeparator(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).separator = s
	}
}

func WithLoaded(f func(k string, v interface{})) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).loaded = f
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			prefix:    "",
			separator: "_",
		}
	}
	return options
}
