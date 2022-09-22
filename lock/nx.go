package lock

import (
	"context"
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

func (nl NxLock) Lock() (ok bool) {
	if !nl.valid {
		return
	}
	ok, _ = nl.ops.redis.SetNX(context.Background(), nl.ops.key, 1, nl.ops.expiration).Result()
	return
}

func (nl NxLock) Unlock() {
	if !nl.valid {
		return
	}
	nl.ops.redis.Del(context.Background(), nl.ops.key)
	return
}
