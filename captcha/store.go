package captcha

import (
	"github.com/mojocn/base64Captcha"
	"time"
)

// redis lua script
const (
	lua string = `
local current = redis.call('GET', KEYS[1])
if current == '' then
    return '';
end
local del = redis.call('DEL', KEYS[1])
if del == 1 then
     return current;
else
     return '';
end
`
)

type Store struct {
	ops      Options
	memory   base64Captcha.Store
	duration time.Duration
}

var memory base64Captcha.Store

func NewStore(options ...func(*Options)) *Store {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	st := &Store{
		ops:      *ops,
		duration: time.Duration(ops.expire) * time.Minute,
	}
	if ops.redis == nil && memory == nil {
		memory = base64Captcha.NewMemoryStore(100, st.duration)
	}
	if st.memory == nil {
		st.memory = memory
	}
	return st
}

func (st Store) Set(id string, value string) (err error) {
	if st.memory != nil {
		err = st.memory.Set(id, value)
	} else {
		_, err = st.ops.redis.Set(st.ops.ctx, st.ops.prefix+id, value, st.duration).Result()
	}
	return
}

func (st Store) Get(id string, clear bool) (rp string) {
	if st.memory != nil {
		rp = st.memory.Get(id, clear)
	} else {
		if !clear {
			rp, _ = st.ops.redis.Get(st.ops.ctx, st.ops.prefix+id).Result()
		} else {
			v, _ := st.ops.redis.Eval(st.ops.ctx, lua, []string{st.ops.prefix + id}).Result()
			if item, ok := v.(string); ok {
				rp = item
			}
		}
	}
	return
}

func (st Store) Verify(id, answer string, clear bool) (rp bool) {
	if st.memory != nil {
		v := st.memory.Get(id, clear)
		if v == answer {
			rp = true
		}
	} else {
		v := st.Get(id, clear)
		if v == answer {
			rp = true
		}
	}
	return
}
