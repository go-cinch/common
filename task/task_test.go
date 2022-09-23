package task

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestTask(t *testing.T) {
	// cron task
	tk1 := New()
	// example 1: run at the 5th minute of each hour
	tk1.Cron(
		WithRunUuid("order1"),
		WithRunName("task1"),
		WithRunExpr("5 * * * ?"),
	)
	// example 2: run at every 8 minute
	tk1.Cron(
		WithRunUuid("order2"),
		WithRunName("task2"),
		WithRunExpr("0/8 * * * ?"),
	)
	// example 3: run at 01 02:00 of month
	tk1.Cron(
		WithRunUuid("order3"),
		WithRunName("task3"),
		WithRunExpr("0 2 1 * ?"),
	)
	// example 4: run at every 10:15 from monday to friday
	tk1.Cron(
		WithRunUuid("order4"),
		WithRunName("task4"),
		WithRunExpr("15 10 ? * MON-FRI"),
	)
	// example 5: run at 10:15 of each month last day
	tk1.Cron(
		WithRunUuid("order5"),
		WithRunName("task5"),
		WithRunExpr("15 10 L * ?"),
	)

	// once task
	tk2 := New(
		WithGroup("task.once"),
		WithHandler(func(ctx context.Context, p Payload) error {
			time.Sleep(time.Minute)
			fmt.Println(ctx)
			return nil
		}),
	)
	tk2.Once(
		WithRunUuid(uuid.NewString()),
		WithRunName("once.task"),
		WithRunAt(time.Now().Add(time.Duration(10)*time.Second)),
		WithRunTimeout(10),
	)

	time.Sleep(time.Minute * 100)
}
