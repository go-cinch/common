package caller

type Options struct {
	skips   []string
	source  bool
	prefix  string
	level   int
	version bool
}

// WithSkip if line contains s, skip it
func WithSkip(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).skips = append(getOptionsOrSetDefault(options).skips, s)
	}
}

// WithSource show common library source code or not
func WithSource(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).source = flag
	}
}

// WithLevel path / count
func WithLevel(i int) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).level = i
	}
}

// WithVersion show library version
func WithVersion(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).version = flag
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			skips:   []string{"gorm.io", "go-kratos", "golang.org/x/sync", "go-cinch/common"},
			source:  false,
			level:   2,
			version: true,
		}
	}
	return options
}
