package worker

import (
	"context"
	"reflect"
	"time"

	"github.com/go-cinch/common/log"
	"github.com/hibiken/asynq"
)

type Options struct {
	group                    string
	redisURI                 string
	redisPeriodKey           string
	retention                int
	maxRetry                 int
	retryDelayFunc           func(n int, e error, t *asynq.Task) time.Duration
	handler                  func(ctx context.Context, p Payload) error
	handlerNeedWorker        func(ctx context.Context, worker Worker, p Payload) error
	callback                 string
	clearArchived            int
	maxArchivedTime          int
	timeout                  int
	delayedTaskCheckInterval time.Duration
	scanTaskInterval         time.Duration
	logLevel                 log.Level
	lockerTTL                time.Duration
	lockerRetryCount         int
	lockerRetryInterval      time.Duration
	streamMaxCount           int // once stream when count > streamMaxCount, overflow data will be removed later
	streamRPS                int // once stream when RPS(request per second) > streamRPS, will sleep more time to process stream queue
}

func WithGroup(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).group = s
	}
}

func WithRedisURI(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).redisURI = s
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

func WithRetryDelayFunc(f func(n int, e error, t *asynq.Task) time.Duration) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).retryDelayFunc = f
	}
}

func WithHandler(fun func(ctx context.Context, p Payload) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).handler = fun
		}
	}
}

func WithHandlerNeedWorker(fun func(ctx context.Context, worker Worker, p Payload) error) func(*Options) {
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

func WithMaxArchivedTime(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).maxArchivedTime = second
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

func WithDelayedTaskCheckInterval(duration time.Duration) func(*Options) {
	return func(options *Options) {
		if duration > 0 {
			getOptionsOrSetDefault(options).delayedTaskCheckInterval = duration
		}
	}
}

func WithScanTaskInterval(duration time.Duration) func(*Options) {
	return func(options *Options) {
		if duration > 0 {
			getOptionsOrSetDefault(options).scanTaskInterval = duration
		}
	}
}

func WithLogLevel(level log.Level) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).logLevel = level
	}
}

func WithLockerTTL(duration time.Duration) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lockerTTL = duration
	}
}

func WithLockerRetryCount(count int) func(*Options) {
	return func(options *Options) {
		if count > 0 {
			getOptionsOrSetDefault(options).lockerRetryCount = count
		}
	}
}

func WithLockerRetryInterval(duration time.Duration) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lockerRetryInterval = duration
	}
}

func WithStreamMaxCount(count int) func(*Options) {
	return func(options *Options) {
		if count > 0 {
			getOptionsOrSetDefault(options).streamMaxCount = count
		}
	}
}

func WithStreamRPS(count int) func(*Options) {
	return func(options *Options) {
		if count > 0 {
			getOptionsOrSetDefault(options).streamRPS = count
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			group:          "task",
			redisURI:       "redis://127.0.0.1:6379/0",
			redisPeriodKey: "period",
			retention:      60,
			maxRetry:       3,
			clearArchived:  300,
			timeout:        10,
			// if u need run seconds expr, must set this param < one period
			// for example, expr is */2 * * * * * *, delayedTaskCheckInterval < 2 * time.Second can work well
			delayedTaskCheckInterval: 5 * time.Second,
			// if u need run seconds expr, must set this param < one period
			scanTaskInterval:    time.Second,
			logLevel:            log.InfoLevel,
			lockerTTL:           time.Minute,
			lockerRetryCount:    40,
			lockerRetryInterval: 25 * time.Millisecond,
			streamMaxCount:      5000,
			streamRPS:           100,
		}
	}
	return options
}

type RunOptions struct {
	uid                 string
	group               string
	payload             string
	expr                string         // only period task, seconds expr: */30 * * * * * *, minutes expr: 0 */5 * * * * *
	in                  *time.Duration // only once task
	at                  *time.Time     // only once task
	now                 bool           // only once task
	retention           int            // only once task
	replace             bool           // only once task
	maxRetry            int
	maxArchivedTime     int
	timeout             int
	lockerTTL           time.Duration
	lockerRetryCount    int
	lockerRetryInterval time.Duration
}

func WithRunUUID(s string) func(*RunOptions) {
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

func WithRunMaxArchivedTime(second int) func(*RunOptions) {
	return func(options *RunOptions) {
		if second > 0 {
			getRunOptionsOrSetDefault(options).maxArchivedTime = second
		}
	}
}

func WithRunLockerTTL(duration time.Duration) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).lockerTTL = duration
	}
}

func WithRunLockerRetryCount(count int) func(*RunOptions) {
	return func(options *RunOptions) {
		if count > 0 {
			getRunOptionsOrSetDefault(options).lockerRetryCount = count
		}
	}
}

func WithRunLockerRetryInterval(duration time.Duration) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).lockerRetryInterval = duration
	}
}

func getRunOptionsOrSetDefault(options *RunOptions) *RunOptions {
	if options == nil {
		return &RunOptions{
			group:               "group",
			timeout:             60,
			lockerTTL:           time.Minute,
			lockerRetryCount:    40,
			lockerRetryInterval: 25 * time.Millisecond,
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
