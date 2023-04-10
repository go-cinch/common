package utils

import (
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func ParseRedisURI(uri string) (client redis.UniversalClient, err error) {
	var opt asynq.RedisConnOpt
	if uri != "" {
		opt, err = asynq.ParseRedisURI(uri)
		if err != nil {
			return
		}
		client = opt.MakeRedisClient().(redis.UniversalClient)
		return
	}
	err = errors.Errorf("invalid redis uri")
	return
}
