package worker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-cinch/common/log"
	"github.com/google/uuid"
)

func TestWorker(*testing.T) {
	// cron worker
	wk1 := New()
	// example 1: run at the 5th minute of each hour
	wk1.Cron(
		context.Background(),
		WithRunUUID("order1"),
		WithRunGroup("task1"),
		WithRunExpr("5 * * * ?"),
	)
	// example 2: run at every 8 minute
	wk1.Cron(
		context.Background(),
		WithRunUUID("order2"),
		WithRunGroup("task2"),
		WithRunExpr("0/8 * * * ?"),
	)
	// example 3: run at 01 02:00 of month
	wk1.Cron(
		context.Background(),
		WithRunUUID("order3"),
		WithRunGroup("task3"),
		WithRunExpr("0 2 1 * ?"),
	)
	// example 4: run at every 10:15 from monday to friday
	wk1.Cron(
		context.Background(),
		WithRunUUID("order4"),
		WithRunGroup("task4"),
		WithRunExpr("15 10 ? * MON-FRI"),
	)
	// example 5: run at 10:15 of each month last day
	wk1.Cron(
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
	wk2.Once(
		context.Background(),
		WithRunUUID(uuid.NewString()),
		WithRunGroup("once.task"),
		WithRunAt(time.Now().Add(time.Duration(10)*time.Second)),
		WithRunTimeout(10),
	)

	time.Sleep(time.Minute * 100)
}

func TestUpdateAndRestoreCronExpr(t *testing.T) {
	const delayTaskUID = "test-delay-cron"
	const advanceTaskUID = "test-advance-cron"
	const oneHourExpr = "0 * * * *"     // run every hour
	const twoHourExpr = "0 */2 * * *"   // run every 2 hours
	const halfHourExpr = "0,30 * * * *" // run every 30 minutes

	fmt.Println("=== Test: UpdateCronExpr and RestoreCronExpr ===")
	fmt.Printf("[%s] Current time\n", time.Now().Format("2006-01-02 15:04:05"))

	var wg sync.WaitGroup
	wg.Add(2)

	// ========================================
	// Goroutine 1: Test delay task (1 hour -> 2 hours -> restore)
	// ========================================
	go func() {
		defer wg.Done()

		fmt.Println("\n========== [Goroutine 1] Delay Task Test ==========")

		wk1 := New(
			WithRedisURI("redis://127.0.0.1:6379/0"),
			WithGroup("test.cron"),
			WithHandlerNeedWorker(func(ctx context.Context, worker Worker, p Payload) error {
				fmt.Printf("[%s] [Delay] Task executed: uid=%s, payload=%s\n",
					time.Now().Format("2006-01-02 15:04:05"), p.UID, p.Payload)
				return nil
			}),
		)
		if wk1.Error != nil {
			fmt.Printf("[Goroutine 1] Failed to create worker: %v\n", wk1.Error)
			return
		}

		// helper function to get task info from redis
		getTaskNext := func(uid string) string {
			t, err := wk1.redis.HGet(context.Background(), wk1.ops.redisPeriodKey, uid).Result()
			if err != nil {
				return "N/A"
			}
			var task periodTask
			task.FromString(t)
			return fmt.Sprintf("exprs=%v, next=%s", task.Exprs, time.Unix(task.Next, 0).Format("2006-01-02 15:04:05"))
		}

		// remove old task if exists
		_ = wk1.Remove(context.Background(), delayTaskUID)

		// create cron task: run every hour
		err := wk1.Cron(
			context.Background(),
			WithRunUUID(delayTaskUID),
			WithRunGroup("test.delay"),
			WithRunExpr(oneHourExpr),
			WithRunPayload(`{"test": "delay-cron"}`),
		)
		if err != nil {
			fmt.Printf("[Goroutine 1] Failed to create cron task: %v\n", err)
			return
		}
		fmt.Printf("[%s] [Delay] Task created: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(delayTaskUID))

		// Test 1: Delay task (1 hour -> 2 hours)
		fmt.Println("\n--- [Delay] Test 1: Delay task (1 hour -> 2 hours) ---")
		time.Sleep(10 * time.Second)
		fmt.Printf("[%s] [Delay] Before update: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(delayTaskUID))
		err = wk1.UpdateCronExpr(context.Background(), delayTaskUID, twoHourExpr)
		if err != nil {
			fmt.Printf("[%s] [Delay] UpdateCronExpr failed: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		} else {
			fmt.Printf("[%s] [Delay] After update:  %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(delayTaskUID))
		}

		// Test 2: Restore task (2 hours -> 1 hour)
		fmt.Println("\n--- [Delay] Test 2: Restore task (2 hours -> 1 hour) ---")
		time.Sleep(30 * time.Second)
		fmt.Printf("[%s] [Delay] Before restore: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(delayTaskUID))
		err = wk1.RestoreCronExpr(context.Background(), delayTaskUID)
		if err != nil {
			fmt.Printf("[%s] [Delay] RestoreCronExpr failed: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		} else {
			fmt.Printf("[%s] [Delay] After restore:  %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(delayTaskUID))
		}

		// keep task running for observation
		time.Sleep(60 * time.Second)

		// cleanup
		_ = wk1.Remove(context.Background(), delayTaskUID)
		fmt.Printf("[%s] [Delay] Task removed\n", time.Now().Format("2006-01-02 15:04:05"))
	}()

	// ========================================
	// Goroutine 2: Test advance task (1 hour -> 30 minutes -> restore)
	// ========================================
	go func() {
		defer wg.Done()

		fmt.Println("\n========== [Goroutine 2] Advance Task Test ==========")

		wk2 := New(
			WithRedisURI("redis://127.0.0.1:6379/0"),
			WithGroup("test.cron"),
			WithHandlerNeedWorker(func(ctx context.Context, worker Worker, p Payload) error {
				fmt.Printf("[%s] [Advance] Task executed: uid=%s, payload=%s\n",
					time.Now().Format("2006-01-02 15:04:05"), p.UID, p.Payload)
				return nil
			}),
		)
		if wk2.Error != nil {
			fmt.Printf("[Goroutine 2] Failed to create worker: %v\n", wk2.Error)
			return
		}

		// helper function to get task info from redis
		getTaskNext := func(uid string) string {
			t, err := wk2.redis.HGet(context.Background(), wk2.ops.redisPeriodKey, uid).Result()
			if err != nil {
				return "N/A"
			}
			var task periodTask
			task.FromString(t)
			return fmt.Sprintf("exprs=%v, next=%s", task.Exprs, time.Unix(task.Next, 0).Format("2006-01-02 15:04:05"))
		}

		// remove old task if exists
		_ = wk2.Remove(context.Background(), advanceTaskUID)

		// create cron task: run every hour
		err := wk2.Cron(
			context.Background(),
			WithRunUUID(advanceTaskUID),
			WithRunGroup("test.advance"),
			WithRunExpr(oneHourExpr),
			WithRunPayload(`{"test": "advance-cron"}`),
		)
		if err != nil {
			fmt.Printf("[Goroutine 2] Failed to create cron task: %v\n", err)
			return
		}
		fmt.Printf("[%s] [Advance] Task created: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(advanceTaskUID))

		// Test 3: Advance task (1 hour -> 30 minutes)
		fmt.Println("\n--- [Advance] Test 3: Advance task (1 hour -> 30 minutes) ---")
		time.Sleep(10 * time.Second)
		fmt.Printf("[%s] [Advance] Before update: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(advanceTaskUID))
		err = wk2.UpdateCronExpr(context.Background(), advanceTaskUID, halfHourExpr)
		if err != nil {
			fmt.Printf("[%s] [Advance] UpdateCronExpr failed: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		} else {
			fmt.Printf("[%s] [Advance] After update:  %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(advanceTaskUID))
		}

		// Test 4: Restore again (30 minutes -> 1 hour)
		fmt.Println("\n--- [Advance] Test 4: Restore task (30 minutes -> 1 hour) ---")
		time.Sleep(30 * time.Second)
		fmt.Printf("[%s] [Advance] Before restore: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(advanceTaskUID))
		err = wk2.RestoreCronExpr(context.Background(), advanceTaskUID)
		if err != nil {
			fmt.Printf("[%s] [Advance] RestoreCronExpr failed: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		} else {
			fmt.Printf("[%s] [Advance] After restore:  %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(advanceTaskUID))
		}

		// keep task running for observation
		time.Sleep(60 * time.Second)

		// cleanup
		_ = wk2.Remove(context.Background(), advanceTaskUID)
		fmt.Printf("[%s] [Advance] Task removed\n", time.Now().Format("2006-01-02 15:04:05"))
	}()

	// ========================================
	// Goroutine 3: Test Cron skip/overwrite behavior
	// ========================================
	wg.Add(1)
	go func() {
		defer wg.Done()

		const skipTaskUID = "test-skip-cron"

		fmt.Println("\n========== [Goroutine 3] Cron Skip/Overwrite Test ==========")

		wk3 := New(
			WithRedisURI("redis://127.0.0.1:6379/0"),
			WithGroup("test.cron"),
			WithHandlerNeedWorker(func(ctx context.Context, worker Worker, p Payload) error {
				fmt.Printf("[%s] [Skip] Task executed: uid=%s, payload=%s\n",
					time.Now().Format("2006-01-02 15:04:05"), p.UID, p.Payload)
				return nil
			}),
		)
		if wk3.Error != nil {
			fmt.Printf("[Goroutine 3] Failed to create worker: %v\n", wk3.Error)
			return
		}

		// helper function to get task info from redis
		getTaskNext := func(uid string) string {
			t, err := wk3.redis.HGet(context.Background(), wk3.ops.redisPeriodKey, uid).Result()
			if err != nil {
				return "N/A"
			}
			var task periodTask
			task.FromString(t)
			return fmt.Sprintf("exprs=%v, originalExprs=%v, next=%s", task.Exprs, task.OriginalExprs, time.Unix(task.Next, 0).Format("2006-01-02 15:04:05"))
		}

		// remove old task if exists
		_ = wk3.Remove(context.Background(), skipTaskUID)

		// Step 1: Create cron task with 1 hour expr
		fmt.Println("\n--- [Skip] Step 1: Create task with 1 hour expr ---")
		err := wk3.Cron(
			context.Background(),
			WithRunUUID(skipTaskUID),
			WithRunGroup("test.skip"),
			WithRunExpr(oneHourExpr),
			WithRunPayload(`{"test": "skip-cron"}`),
		)
		if err != nil {
			fmt.Printf("[Goroutine 3] Failed to create cron task: %v\n", err)
			return
		}
		fmt.Printf("[%s] [Skip] Task created: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))

		// Step 2: Update to 2 hours (this sets originalExpr)
		fmt.Println("\n--- [Skip] Step 2: Update to 2 hours (sets originalExpr) ---")
		time.Sleep(5 * time.Second)
		fmt.Printf("[%s] [Skip] Before update: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))
		err = wk3.UpdateCronExpr(context.Background(), skipTaskUID, twoHourExpr)
		if err != nil {
			fmt.Printf("[%s] [Skip] UpdateCronExpr failed: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		} else {
			fmt.Printf("[%s] [Skip] After update:  %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))
		}

		// Step 3: Call Cron again with 1 hour expr (should be skipped because originalExpr is set)
		fmt.Println("\n--- [Skip] Step 3: Call Cron again with 1 hour (should SKIP) ---")
		time.Sleep(5 * time.Second)
		fmt.Printf("[%s] [Skip] Before Cron: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))
		err = wk3.Cron(
			context.Background(),
			WithRunUUID(skipTaskUID),
			WithRunGroup("test.skip"),
			WithRunExpr(oneHourExpr),
			WithRunPayload(`{"test": "skip-cron"}`),
		)
		fmt.Printf("[%s] [Skip] After Cron:  %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))

		// Step 4: Restore to original (clears originalExpr)
		fmt.Println("\n--- [Skip] Step 4: Restore (clears originalExpr) ---")
		time.Sleep(5 * time.Second)
		fmt.Printf("[%s] [Skip] Before restore: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))
		err = wk3.RestoreCronExpr(context.Background(), skipTaskUID)
		if err != nil {
			fmt.Printf("[%s] [Skip] RestoreCronExpr failed: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		} else {
			fmt.Printf("[%s] [Skip] After restore:  %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))
		}

		// Step 5: Call Cron again with 30 min expr (should OVERWRITE because originalExpr is empty)
		fmt.Println("\n--- [Skip] Step 5: Call Cron with 30 min (should OVERWRITE) ---")
		time.Sleep(5 * time.Second)
		fmt.Printf("[%s] [Skip] Before Cron: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))
		err = wk3.Cron(
			context.Background(),
			WithRunUUID(skipTaskUID),
			WithRunGroup("test.skip"),
			WithRunExpr(halfHourExpr),
			WithRunPayload(`{"test": "skip-cron"}`),
		)
		fmt.Printf("[%s] [Skip] After Cron:  %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(skipTaskUID))

		// keep task running for observation
		time.Sleep(30 * time.Second)

		// cleanup
		_ = wk3.Remove(context.Background(), skipTaskUID)
		fmt.Printf("[%s] [Skip] Task removed\n", time.Now().Format("2006-01-02 15:04:05"))
	}()

	// ========================================
	// Goroutine 4: Test UpdateCronExpr inside handler
	// ========================================
	wg.Add(1)
	go func() {
		defer wg.Done()

		const handlerTaskUID = "test-handler-update-cron"
		const tenSecExpr = "*/10 * * * * * *" // run every 10 seconds
		const fiveHourExpr = "0 */5 * * *"    // run every 5 hours

		fmt.Println("\n========== [Goroutine 4] Handler Update Test ==========")

		var once sync.Once
		var wk4 *Worker

		wk4 = New(
			WithRedisURI("redis://127.0.0.1:6379/0"),
			WithGroup("test.cron"),
			WithHandlerNeedWorker(func(ctx context.Context, worker Worker, p Payload) error {
				fmt.Printf("[%s] [Handler] Task executed: uid=%s\n",
					time.Now().Format("2006-01-02 15:04:05"), p.UID)

				// only update once
				once.Do(func() {
					go func() {
						fmt.Printf("[%s] [Handler] Will update expr in 5 seconds...\n",
							time.Now().Format("2006-01-02 15:04:05"))
						time.Sleep(5 * time.Second)
						err := worker.UpdateCronExpr(context.Background(), p.UID, fiveHourExpr)
						if err != nil {
							fmt.Printf("[%s] [Handler] UpdateCronExpr failed: %v\n",
								time.Now().Format("2006-01-02 15:04:05"), err)
						} else {
							fmt.Printf("[%s] [Handler] UpdateCronExpr success: changed to 5 hours\n",
								time.Now().Format("2006-01-02 15:04:05"))
						}
					}()
				})
				return nil
			}),
		)
		if wk4.Error != nil {
			fmt.Printf("[Goroutine 4] Failed to create worker: %v\n", wk4.Error)
			return
		}

		// helper function to get task info from redis
		getTaskNext := func(uid string) string {
			t, err := wk4.redis.HGet(context.Background(), wk4.ops.redisPeriodKey, uid).Result()
			if err != nil {
				return "N/A"
			}
			var task periodTask
			task.FromString(t)
			return fmt.Sprintf("exprs=%v, next=%s", task.Exprs, time.Unix(task.Next, 0).Format("2006-01-02 15:04:05"))
		}

		// remove old task if exists
		_ = wk4.Remove(context.Background(), handlerTaskUID)

		// Create cron task: run every 10 seconds
		fmt.Println("\n--- [Handler] Create task with 10 second interval ---")
		err := wk4.Cron(
			context.Background(),
			WithRunUUID(handlerTaskUID),
			WithRunGroup("test.handler"),
			WithRunExpr(tenSecExpr),
			WithRunPayload(`{"test": "handler-update"}`),
		)
		if err != nil {
			fmt.Printf("[Goroutine 4] Failed to create cron task: %v\n", err)
			return
		}
		fmt.Printf("[%s] [Handler] Task created: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(handlerTaskUID))

		// worker already started in New(), just wait and observe
		for i := 0; i < 12; i++ {
			time.Sleep(10 * time.Second)
			fmt.Printf("[%s] [Handler] Current state: %s\n", time.Now().Format("2006-01-02 15:04:05"), getTaskNext(handlerTaskUID))
		}

		// cleanup
		_ = wk4.Remove(context.Background(), handlerTaskUID)
		fmt.Printf("[%s] [Handler] Task removed\n", time.Now().Format("2006-01-02 15:04:05"))
	}()

	// wait for all goroutines to complete
	wg.Wait()
	fmt.Printf("\n[%s] Test completed, all tasks removed\n", time.Now().Format("2006-01-02 15:04:05"))
}

// TestCronSingleAndMultipleExpressions tests creating cron tasks with single and multiple expressions
func TestCronSingleAndMultipleExpressions(t *testing.T) {
	// Create worker
	wk := New(
		WithRedisURI("redis://127.0.0.1:6379/0"),
		WithGroup("test-cron-exprs"),
	)

	if wk.Error != nil {
		t.Fatalf("Failed to create worker: %v", wk.Error)
	}

	ctx := context.Background()
	singleTaskUID := "test-single-expr"
	multiTaskUID := "test-multi-expr"

	// Helper function to get task info from Redis
	getTaskInfo := func(uid string) string {
		t, err := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, uid).Result()
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		var task periodTask
		task.FromString(t)
		return fmt.Sprintf("exprs=%v, next=%s, processed=%d",
			task.Exprs,
			time.Unix(task.Next, 0).Format("2006-01-02 15:04:05"),
			task.Processed)
	}

	// Clean up any existing tasks
	_ = wk.Remove(ctx, singleTaskUID)
	_ = wk.Remove(ctx, multiTaskUID)

	t.Log("========== Test 1: Create Cron with Single Expression ==========")

	// Create cron task with single expression
	err := wk.Cron(
		ctx,
		WithRunUUID(singleTaskUID),
		WithRunGroup("test-cron-exprs"),
		WithRunExpr("0 0 10 * * * *"), // every day at 10:00 AM
		WithRunPayload(`{"task": "single"}`),
	)

	if err != nil {
		t.Errorf("Failed to create cron task with single expression: %v", err)
	} else {
		t.Logf("✓ Single expression task created successfully")
		t.Logf("  Task info: %s", getTaskInfo(singleTaskUID))
	}

	time.Sleep(2 * time.Second)

	t.Log("\n========== Test 2: Create Cron with Multiple Expressions ==========")

	// Create cron task with multiple expressions
	multiExprs := []string{
		"0 30 18 * * * *", // every day at 6:30 PM
		"0 30 19 * * * *", // every day at 7:30 PM
		"0 30 20 * * * *", // every day at 8:30 PM
		"0 30 21 * * * *", // every day at 9:30 PM
		"0 30 22 * * * *", // every day at 10:30 PM
	}

	err = wk.Cron(
		ctx,
		WithRunUUID(multiTaskUID),
		WithRunGroup("test-cron-exprs"),
		WithRunExpr(multiExprs...),
		WithRunPayload(`{"task": "multiple"}`),
	)

	if err != nil {
		t.Errorf("Failed to create cron task with multiple expressions: %v", err)
	} else {
		t.Logf("✓ Multiple expressions task created successfully")
		t.Logf("  Task info: %s", getTaskInfo(multiTaskUID))
	}

	time.Sleep(2 * time.Second)

	t.Log("\n========== Test 3: Verify Both Tasks Exist ==========")

	// Verify single expression task
	singleInfo := getTaskInfo(singleTaskUID)
	t.Logf("✓ Single expression task: %s", singleInfo)

	// Verify multiple expressions task
	multiInfo := getTaskInfo(multiTaskUID)
	t.Logf("✓ Multiple expressions task: %s", multiInfo)

	t.Log("\n========== Test 4: Update Task ==========")

	// Update multiple expressions task to use single expression
	err = wk.UpdateCronExpr(
		ctx,
		multiTaskUID,
		"0 0 17 * * * *", // every day at 4:00 PM
	)

	if err != nil {
		t.Errorf("Failed to update multiple task to single expression: %v", err)
	} else {
		t.Logf("✓ Updated multiple task to single expression")
		t.Logf("  Updated task info: %s", getTaskInfo(multiTaskUID))
	}

	time.Sleep(2 * time.Second)

	t.Log("\n========== Test 6: Clean Up ==========")

	// Clean up
	err = wk.Remove(ctx, singleTaskUID)
	if err != nil {
		t.Errorf("Failed to remove single task: %v", err)
	} else {
		t.Logf("✓ Removed single expression task")
	}

	err = wk.Remove(ctx, multiTaskUID)
	if err != nil {
		t.Errorf("Failed to remove multiple task: %v", err)
	} else {
		t.Logf("✓ Removed multiple expressions task")
	}

	t.Log("\n========== Test Complete ==========")
}

// TestRemoveCancelsSlowTask verifies that calling Remove on a running task
// sends a cancel signal to the task's context.
func TestRemoveCancelsSlowTask(t *testing.T) {
	ctx := context.Background()
	uid := "test-cancel-slow-task-" + uuid.NewString()

	startedCh := make(chan struct{}, 1)
	cancelledCh := make(chan struct{}, 1)
	doneCh := make(chan struct{}, 1)

	wk := New(
		WithRedisURI("redis://127.0.0.1:6379/0"),
		WithGroup("test.cancel"),
		WithHandlerNeedWorker(func(ctx context.Context, worker Worker, p Payload) error {
			// mark that task processing has started (this worker is dedicated to this test)
			select {
			case startedCh <- struct{}{}:
			default:
			}

			log.Info("task is running")
			// simulate a long-running task that is sensitive to context cancellation
			select {
			case <-ctx.Done():
				// notify test that we observed cancellation
				select {
				case cancelledCh <- struct{}{}:
					log.Info("task detected context cancellation")
				default:
				}
				return ctx.Err()
			case <-time.After(30 * time.Second):
				// if cancel doesn't happen, this branch would eventually fire
				select {
				case doneCh <- struct{}{}:
				default:
				}
				return nil
			}
		}),
	)
	if wk.Error != nil {
		t.Fatalf("failed to create worker: %v", wk.Error)
	}

	// ensure no leftover task with the same uid
	_ = wk.Remove(ctx, uid)

	// enqueue a once task that should start processing soon and run for a long time unless cancelled
	err := wk.Once(
		ctx,
		WithRunUUID(uid),
		WithRunGroup("test.cancel"),
		WithRunPayload(`{"task":"slow"}`),
		WithRunTimeout(60),
		WithRunNow(true),
	)
	if err != nil {
		t.Fatalf("failed to enqueue once task: %v", err)
	}

	// wait for handler to start processing the task
	select {
	case <-startedCh:
		// ok
	case <-time.After(60 * time.Second):
		t.Fatalf("task did not start processing in time; ensure Redis is running on 127.0.0.1:6379")
	}

	log.Info("wait 5s send cancel signal")
	time.Sleep(5 * time.Second)

	// call Remove while the task is running to trigger CancelProcessing
	removeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := wk.Remove(removeCtx, uid); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	// verify that the task observed context cancellation instead of completing normally
	select {
	case <-cancelledCh:
		// expected path: context was cancelled
	case <-time.After(15 * time.Second):
		t.Fatalf("task context was not cancelled within timeout after Remove")
	}

	// best-effort check that the normal-completion branch did not run
	select {
	case <-doneCh:
		t.Fatalf("slow task completed normally instead of being cancelled")
	default:
	}
}
