package idempotence

import (
	"context"
	"fmt"
	"github.com/go-cinch/common/log"
	"github.com/google/uuid"
)

// redis lua script(read => delete => get delete flag)
const (
	lua string = `
local current = redis.call('GET', KEYS[1])
if current == false then
    return '-1';
end
local del = redis.call('DEL', KEYS[1])
if del == 1 then
     return '1';
else
     return '0';
end
`
)

type Idempotence struct {
	ops Options
}

func New(options ...func(*Options)) *Idempotence {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return &Idempotence{ops: *ops}
}

func (i *Idempotence) Token(ctx context.Context) (token string) {
	token = uuid.NewString()
	if i.ops.redis != nil {
		i.ops.redis.Set(ctx, fmt.Sprintf("%s_%s", i.ops.prefix, token), true, i.ops.expiration)
	} else {
		log.WithContext(ctx).Warn("please enable redis, otherwise the idempotence is invalid")
	}
	return
}

func (i *Idempotence) Check(ctx context.Context, token string) (pass bool) {
	if i.ops.redis != nil {
		res, err := i.ops.redis.Eval(ctx, lua, []string{fmt.Sprintf("%s_%s", i.ops.prefix, token)}).Result()
		if err != nil || res != "1" {
			return
		}
	} else {
		log.WithContext(ctx).Warn("please enable redis, otherwise the idempotence is invalid")
	}
	pass = true
	return
}
