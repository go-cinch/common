package idempotent

import "github.com/redis/go-redis/v9"

type Options struct {
	redis  redis.UniversalClient
	prefix string
	expire int
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		if prefix != "" {
			getOptionsOrSetDefault(options).prefix = prefix
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

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			prefix: "idempotent",
			expire: 60,
		}
	}
	return options
}
