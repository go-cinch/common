package worker

import (
	"context"
	"reflect"
	"time"
)

type Options struct {
	group             string
	redisUri          string
	redisPeriodKey    string
	retention         int
	maxRetry          int
	handler           func(ctx context.Context, p Payload) error
	handlerNeedWorker func(worker Worker, ctx context.Context, p Payload) error
	callback          string
	clearArchived     int
	timeout           int
}

func WithGroup(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).group = s
	}
}

func WithRedisUri(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).redisUri = s
	}
}

func WithRedisPeriodKey(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).redisPeriodKey = s
	}
}

func WithRetention(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).retention = second
		}
	}
}

func WithMaxRetry(count int) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).maxRetry = count
	}
}

func WithHandler(fun func(ctx context.Context, p Payload) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).handler = fun
		}
	}
}

func WithHandlerNeedWorker(fun func(worker Worker, ctx context.Context, p Payload) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).handlerNeedWorker = fun
		}
	}
}

func WithCallback(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).callback = s
	}
}

func WithClearArchived(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).clearArchived = second
		}
	}
}

func WithTimeout(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).timeout = second
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			group:          "task",
			redisUri:       "redis://127.0.0.1:6379/0",
			redisPeriodKey: "period",
			retention:      60,
			maxRetry:       3,
			clearArchived:  300,
			timeout:        10,
		}
	}
	return options
}

type RunOptions struct {
	uid       string
	group     string
	payload   string
	expr      string          // only period task
	in        *time.Duration  // only once task
	at        *time.Time      // only once task
	now       bool            // only once task
	retention int             // only once task
	replace   bool            // only once task
	ctx       context.Context // only once task
	maxRetry  int
	timeout   int
}

func WithRunUuid(s string) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).uid = s
	}
}

func WithRunGroup(s string) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).group = s
	}
}

func WithRunPayload(s string) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).payload = s
	}
}

func WithRunExpr(s string) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).expr = s
	}
}

func WithRunIn(in time.Duration) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).in = &in
	}
}

func WithRunAt(at time.Time) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).at = &at
	}
}

func WithRunNow(flag bool) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).now = flag
	}
}

func WithRunRetention(second int) func(*RunOptions) {
	return func(options *RunOptions) {
		if second > 0 {
			getRunOptionsOrSetDefault(options).retention = second
		}
	}
}

// WithRunReplace remove old one and create new one when uid repeat
func WithRunReplace(flag bool) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).replace = flag
	}
}

func WithRunCtx(ctx context.Context) func(*RunOptions) {
	return func(options *RunOptions) {
		if !interfaceIsNil(ctx) {
			getRunOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithRunMaxRetry(count int) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).maxRetry = count
	}
}

func WithRunTimeout(second int) func(*RunOptions) {
	return func(options *RunOptions) {
		if second > 0 {
			getRunOptionsOrSetDefault(options).timeout = second
		}
	}
}

func getRunOptionsOrSetDefault(options *RunOptions) *RunOptions {
	if options == nil {
		return &RunOptions{
			group:   "group",
			timeout: 60,
		}
	}
	return options
}

func interfaceIsNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return i == nil
}
