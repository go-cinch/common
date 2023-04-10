package bloom

import "github.com/redis/go-redis/v9"

type Options struct {
	redis   redis.UniversalClient
	key     string
	expire  int
	hash    []func(str string) uint64
	timeout int
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithKey(key string) func(*Options) {
	return func(options *Options) {
		if key != "" {
			getOptionsOrSetDefault(options).key = key
		}
	}
}

func WithExpire(min int) func(*Options) {
	return func(options *Options) {
		if min > 0 {
			getOptionsOrSetDefault(options).expire = min
		}
	}
}

func WithHash(f ...func(str string) uint64) func(*Options) {
	return func(options *Options) {
		if len(f) > 0 {
			getOptionsOrSetDefault(options).hash = append(getOptionsOrSetDefault(options).hash, f...)
		}
	}
}

func WithTimeout(seconds int) func(*Options) {
	return func(options *Options) {
		if seconds > 0 {
			getOptionsOrSetDefault(options).timeout = seconds
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			key:     "bloom",
			expire:  5,
			hash:    []func(str string) uint64{BKDRHash, SDBMHash, DJBHash},
			timeout: 3,
		}
	}
	return options
}
