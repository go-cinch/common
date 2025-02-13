package idempotent

import (
	"context"
	"strings"

	"github.com/go-cinch/common/log"
)

// redis lua script(read => not exist set or exist return)
const (
	lua string = `
local current = redis.call('GET', KEYS[1])
if current ~= false then
    return '0';
end
redis.call('SET', KEYS[1], ARGV[1])
redis.call('EXPIRE', KEYS[1], ARGV[2])
return '1';
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

func (i *Idempotent) Check(ctx context.Context, token string) (pass bool) {
	if i.ops.redis == nil {
		log.WithContext(ctx).Warn("please enable redis, otherwise the idempotent is invalid")
		return
	}
	res, err := i.ops.redis.Eval(
		ctx,
		lua,
		[]string{strings.Join([]string{i.ops.prefix, token}, "_")},
		"1",
		i.ops.expire,
	).Result()
	pass = err == nil && res == "1"
	return
}
