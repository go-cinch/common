package idempotent

import (
	"github.com/go-redis/redis/v8"
	"time"
)

type Options struct {
	redis      redis.UniversalClient
	prefix     string
	expiration time.Duration
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

func WithExpiration(hours int) func(*Options) {
	return func(options *Options) {
		if hours > 0 {
			getOptionsOrSetDefault(options).expiration = time.Duration(hours) * time.Hour
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			prefix:     "idempotence",
			expiration: 24 * time.Hour,
		}
	}
	return options
}
