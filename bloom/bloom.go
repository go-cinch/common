package bloom

import (
	"context"
	"strconv"
	"time"
)

// redis lua script
const (
	luaSet string = `
for _, offset in ipairs(ARGV) do
	redis.call('SETBIT', KEYS[1], offset, 1)
end
`
	luaGet = `
for _, offset in ipairs(ARGV) do
	if tostring(redis.call('GETBIT', KEYS[1], offset)) == '0' then
			return '0'
		end
	end
return '1'
`
)

type Bloom struct {
	ops Options
}

func New(options ...func(*Options)) *Bloom {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	b := &Bloom{
		ops: *ops,
	}
	return b
}

func (b Bloom) Add(str ...string) (err error) {
	ctx := b.getDefaultTimeoutCtx()
	p := b.ops.redis.Pipeline()
	for _, item := range str {
		_, err = p.Eval(ctx, luaSet, []string{b.ops.key}, b.offset(item)).Result()
		if err != nil {
			return
		}
	}
	p.Expire(ctx, b.ops.key, time.Duration(b.ops.expire)*time.Minute)
	_, err = p.Exec(ctx)
	return
}

func (b Bloom) Exist(str string) (ok bool) {
	res, err := b.ops.redis.Eval(b.getDefaultTimeoutCtx(), luaGet, []string{b.ops.key}, b.offset(str)).Result()
	if err != nil || res != "1" {
		return
	}
	ok = true
	return
}

func (b Bloom) Flush() {
	b.ops.redis.Del(b.getDefaultTimeoutCtx(), b.ops.key)
	return
}

func (b Bloom) offset(str string) (list []string) {
	list = make([]string, 0)
	for _, f := range b.ops.hash {
		offset := f(str)
		list = append(list, strconv.FormatUint(offset, 10))
	}
	return
}

func (b Bloom) getDefaultTimeoutCtx() context.Context {
	c, _ := context.WithTimeout(context.Background(), time.Duration(b.ops.timeout)*time.Second)
	return c
}
