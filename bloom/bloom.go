package bloom

import (
	"context"
	"fmt"
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

func (b *Bloom) Add(str ...string) (err error) {
	ctx := context.Background()
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

func (b *Bloom) Exist(str string) (ok bool) {
	res, err := b.ops.redis.Eval(context.Background(), luaGet, []string{b.ops.key}, b.offset(str)).Result()
	if err != nil || res != "1" {
		return
	}
	ok = true
	return
}

func (b *Bloom) offset(str string) (list []string) {
	list = make([]string, 0)
	for _, f := range b.ops.hash {
		offset := f(str)
		list = append(list, fmt.Sprintf("%d", offset))
	}
	return
}
