package id

type Options struct {
	chars []rune
	n1    int
	n2    int
	l     int
	salt  uint64
}

func WithChars(arr []rune) func(*Options) {
	return func(options *Options) {
		if len(arr) > 0 {
			getOptionsOrSetDefault(options).chars = arr
		}
	}
}

func WithN1(n int) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).n1 = n
	}
}

func WithN2(n int) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).n2 = n
	}
}

func WithL(l int) func(*Options) {
	return func(options *Options) {
		if l > 0 {
			getOptionsOrSetDefault(options).l = l
		}
	}
}

func WithSalt(salt uint64) func(*Options) {
	return func(options *Options) {
		if salt > 0 {
			getOptionsOrSetDefault(options).salt = salt
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			// base string set, remove 0,1,I,O,U,Z
			chars: []rune{
				'2', '3', '4', '5', '6',
				'7', '8', '9', 'A', 'B',
				'C', 'D', 'E', 'F', 'G',
				'H', 'J', 'K', 'L', 'M',
				'N', 'P', 'Q', 'R', 'S',
				'T', 'V', 'W', 'X', 'Y',
			},
			// n1 / len(chars)=30 cop rime
			n1: 17,
			// n2 / l cop rime
			n2: 5,
			// code length
			l: 8,
			// random number
			salt: 123567369,
		}
	}
	return options
}
