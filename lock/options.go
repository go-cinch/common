package lock

import (
	"github.com/go-redis/redis/v8"
	"time"
)

type NxLockOptions struct {
	redis      redis.UniversalClient
	key        string
	expiration time.Duration
}

func WithNxLockRedis(rd redis.UniversalClient) func(*NxLockOptions) {
	return func(options *NxLockOptions) {
		if rd != nil {
			getNxLockOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithNxLockKey(key string) func(*NxLockOptions) {
	return func(options *NxLockOptions) {
		if key != "" {
			getNxLockOptionsOrSetDefault(options).key = key
		}
	}
}

func WithNxLockExpiration(seconds int) func(*NxLockOptions) {
	return func(options *NxLockOptions) {
		if seconds > 0 {
			getNxLockOptionsOrSetDefault(options).expiration = time.Duration(seconds) * time.Second
		}
	}
}

func getNxLockOptionsOrSetDefault(options *NxLockOptions) *NxLockOptions {
	if options == nil {
		return &NxLockOptions{
			key:        "nx.lock",
			expiration: time.Minute,
		}
	}
	return options
}
