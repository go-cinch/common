package idempotent

import (
	"context"
	"strings"
	"time"

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
end
return '0';
`
)

type Idempotent struct {
	ops Options
}

func New(options ...func(*Options)) *Idempotent {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return &Idempotent{ops: *ops}
}

func (i *Idempotent) Token(ctx context.Context) (token string) {
	token = uuid.NewString()
	if i.ops.redis != nil {
		i.ops.redis.Set(ctx, strings.Join([]string{i.ops.prefix, token}, "_"), true, time.Duration(i.ops.expire)*time.Minute)
	} else {
		log.WithContext(ctx).Warn("please enable redis, otherwise the idempotent is invalid")
	}
	return
}

func (i *Idempotent) Check(ctx context.Context, token string) (pass bool) {
	if i.ops.redis != nil {
		res, err := i.ops.redis.Eval(ctx, lua, []string{strings.Join([]string{i.ops.prefix, token}, "_")}).Result()
		if err != nil || res != "1" {
			return
		}
	} else {
		log.WithContext(ctx).Warn("please enable redis, otherwise the idempotent is invalid")
	}
	pass = true
	return
}
