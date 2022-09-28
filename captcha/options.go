package captcha

import (
	"context"
	"github.com/go-redis/redis/v8"
	"reflect"
)

type Options struct {
	ctx    context.Context
	redis  redis.UniversalClient
	prefix string
	expire int
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !interfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
		}
	}
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
		getOptionsOrSetDefault(options).prefix = prefix
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
			ctx:    context.Background(),
			prefix: "captcha_",
			expire: 5,
		}
	}
	return options
}

func interfaceIsNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return i == nil
}
