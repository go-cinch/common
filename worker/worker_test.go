package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWorker(*testing.T) {
	// cron worker
	wk1 := New()
	// example 1: run at the 5th minute of each hour
	_ = wk1.Cron(
		context.Background(),
		WithRunUUID("order1"),
		WithRunGroup("task1"),
		WithRunExpr("5 * * * ?"),
	)
	// example 2: run at every 8 minute
	_ = wk1.Cron(
		context.Background(),
		WithRunUUID("order2"),
		WithRunGroup("task2"),
		WithRunExpr("0/8 * * * ?"),
	)
	// example 3: run at 01 02:00 of month
	_ = wk1.Cron(
		context.Background(),
		WithRunUUID("order3"),
		WithRunGroup("task3"),
		WithRunExpr("0 2 1 * ?"),
	)
	// example 4: run at every 10:15 from monday to friday
	_ = wk1.Cron(
		context.Background(),
		WithRunUUID("order4"),
		WithRunGroup("task4"),
		WithRunExpr("15 10 ? * MON-FRI"),
	)
	// example 5: run at 10:15 of each month last day
	_ = wk1.Cron(
		context.Background(),
		WithRunUUID("order5"),
		WithRunGroup("task5"),
		WithRunExpr("15 10 L * ?"),
	)

	// once worker
	wk2 := New(
		WithGroup("task.once"),
		WithHandler(func(ctx context.Context, _ Payload) error {
			time.Sleep(time.Minute)
			fmt.Println(ctx)
			return nil
		}),
	)
	_ = wk2.Once(
		context.Background(),
		WithRunUUID(uuid.NewString()),
		WithRunGroup("once.task"),
		WithRunAt(time.Now().Add(time.Duration(10)*time.Second)),
		WithRunTimeout(10),
	)

	ch := make(chan struct{})
	<-ch
}
