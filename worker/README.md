# Worker

distributed async task worker based on [asynq](https://github.com/hibiken/asynq).

## Usage

```bash
go get -u github.com/go-cinch/common/worker
```

```go
import (
	"context"
	"fmt"

	"github.com/go-cinch/common/worker"
)

func main() {
	wk := worker.New(
		// redis://[:password@]host[:port][/dbnumber]
		// worker.WithRedisURI("redis://:password@127.0.0.1:6379/0"),
		worker.WithRedisURI("redis://127.0.0.1:6379/0"),
		worker.WithHandler(process),
	)
	err := wk.Error
	if err != nil {
		panic(err)
	}

	// 1. cron task
	_ = wk.Cron(
		context.Background(),
		worker.WithRunUUID("order1"),
		worker.WithRunGroup("task1"),
		worker.WithRunExpr("0/1 * * * ?"),
	)

	// 2. once task
	_ = wk.Once(
		context.Background(),
		worker.WithRunUUID("order2"),
		worker.WithRunGroup("task2"),
		worker.WithRunNow(true),
	)

	ch := make(chan struct{})
	<-ch
}

func process(ctx context.Context, p worker.Payload) (err error) {
	switch p.Group {
	case "task1":
		fmt.Println(ctx, p.UID)
	case "task2":
		fmt.Println(ctx, p.UID)
	}
	return
}
```

## Options

### WorkerOptions

- `WithGroup` - group name, default task
- `WithRedisURI` - redis uri, default redis://127.0.0.1:6379/0
- `WithRedisPeriodKey` - cron task cache key
- `WithRetention` - success task store time, default 60s, if this option is provided, the task will be stored as a
  completed task after successful processing
- `WithMaxRetry` - max retry count when task has error, default 3
- `WithHandler` - callback handler
- `WithCallback` - http callback uri
- `WithClearArchived` - clear archived task internal, default 300s
- `WithTimeout` - task timeout, default 10s

### RunOptions

#### Cron

cron task, can be executed multiple times

- `WithRunUUID` - task unique id
- `WithRunGroup` - group prefix, default group
- `WithRunPayload` - task payload
- `WithRunExpr` - cron expr, mini is one minute, refer to [gorhill/cronexpr](https://github.com/gorhill/cronexpr)
- `WithRunMaxRetry` - max retry count when task has error
- `WithRunTimeout` - task timeout, default 60

#### Once

once task, execute only once

- `WithRunUUID` - task unique id
- `WithRunGroup` - group prefix, default group
- `WithRunPayload` - task payload
- `WithRunMaxRetry` - max retry count when task has error
- `WithRunTimeout` - task timeout, default 60
- `WithRunCtx` - context
- `WithRunIn` - run in xxx seconds
- `WithRunAt` - run at
- `WithRunNow` - run now
- `WithRunRetention` - success task store time
- `WithRunReplace` - remove old one and create new one when uid repeat, default false
