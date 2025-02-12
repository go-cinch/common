package stream

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	rds      redis.UniversalClient
	key      string
	group    string
	consumer string
	expire   time.Duration
}

func WithRDS(rds redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).rds = rds
	}
}

func WithKey(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).key = s
	}
}

func WithGroup(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).group = s
	}
}

func WithConsumer(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).consumer = s
	}
}

func WithExpire(duration time.Duration) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).expire = duration
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			expire: 86400 * time.Second,
		}
	}
	return options
}
