package nx

import (
	"context"
	"errors"
	"time"
)

type Nx struct {
	ops   Options
	valid bool
}

func New(options ...func(*Options)) (nx *Nx) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	nx = &Nx{
		ops:   *ops,
		valid: ops.redis != nil,
	}
	return
}

func (nx Nx) MustLock(c ...context.Context) (err error) {
	if !nx.valid {
		return
	}
	ctx := context.Background()
	if len(c) > 0 {
		ctx = c[0]
	}
	var retry int
	for {
		ok, e := nx.ops.redis.SetNX(ctx, nx.ops.key, 1, time.Duration(nx.ops.expire)*time.Second).Result()
		if errors.Is(e, context.DeadlineExceeded) || errors.Is(e, context.Canceled) || (e != nil && e.Error() == "redis: connection pool timeout") {
			err = e
			return
		}
		if ok {
			break
		}
		time.Sleep(25 * time.Millisecond)
		retry++
		if retry > 400 {
			err = errors.New("lock timeout")
			return
		}
	}
	return
}

func (nx Nx) Lock(c ...context.Context) (ok bool) {
	if !nx.valid {
		return
	}
	ctx := context.Background()
	if len(c) > 0 {
		ctx = c[0]
	}
	ok, _ = nx.ops.redis.SetNX(ctx, nx.ops.key, 1, time.Duration(nx.ops.expire)*time.Second).Result()
	return
}

func (nx Nx) Unlock(ctx ...context.Context) {
	if !nx.valid {
		return
	}
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}
	nx.ops.redis.Del(c, nx.ops.key)
	return
}
