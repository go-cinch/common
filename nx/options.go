package nx

import (
	"github.com/go-redis/redis/v8"
)

type Options struct {
	redis  redis.UniversalClient
	key    string
	expire int
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

func WithExpire(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).expire = second
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			key:    "nx.lock",
			expire: 60,
		}
	}
	return options
}
