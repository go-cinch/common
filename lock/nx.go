package lock

import (
	"context"
	"errors"
	"time"
)

type NxLock struct {
	ops   NxLockOptions
	valid bool
}

func NewNxLock(options ...func(*NxLockOptions)) (lock *NxLock) {
	ops := getNxLockOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	lock = &NxLock{
		ops:   *ops,
		valid: ops.redis != nil,
	}
	return
}

func (nl NxLock) MustLock(c ...context.Context) (err error) {
	if !nl.valid {
		return
	}
	ctx := context.Background()
	if len(c) > 0 {
		ctx = c[0]
	}
	var retry int
	for {
		ok, e := nl.ops.redis.SetNX(ctx, nl.ops.key, 1, nl.ops.expiration).Result()
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

func (nl NxLock) Lock(c ...context.Context) (ok bool) {
	if !nl.valid {
		return
	}
	ctx := context.Background()
	if len(c) > 0 {
		ctx = c[0]
	}
	ok, _ = nl.ops.redis.SetNX(ctx, nl.ops.key, 1, nl.ops.expiration).Result()
	return
}

func (nl NxLock) Unlock(ctx ...context.Context) {
	if !nl.valid {
		return
	}
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}
	nl.ops.redis.Del(c, nl.ops.key)
	return
}
