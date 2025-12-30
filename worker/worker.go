package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-cinch/common/log"
	"github.com/go-cinch/common/queue/stream"
	"github.com/golang-module/carbon/v2"
	"github.com/google/uuid"
	"github.com/gorhill/cronexpr"
	"github.com/hibiken/asynq"
	"github.com/paulbellamy/ratecounter"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Worker struct {
	ops           Options
	redis         redis.UniversalClient
	redisOpt      asynq.RedisConnOpt
	locker        *redislock.Client
	client        *asynq.Client
	inspector     *asynq.Inspector
	stream        *stream.Stream
	streamLimiter *ratecounter.RateCounter
	Error         error
}

type periodTask struct {
	Exprs           []string `json:"exprs"`         // multiple cron expressions
	OriginalExprs   []string `json:"originalExprs"` // original cron exprs for restore
	Group           string   `json:"group"`
	UID             string   `json:"uid"`
	Payload         string   `json:"payload"`
	Next            int64    `json:"next"`      // next schedule unix timestamp
	Processed       int64    `json:"processed"` // run times
	MaxRetry        int      `json:"maxRetry"`
	MaxArchivedTime int      `json:"maxArchivedTime"`
	Timeout         int      `json:"timeout"`
}

func (p *periodTask) String() (str string) {
	bs, _ := json.Marshal(p)
	str = string(bs)
	return
}

func (p *periodTask) FromString(str string) {
	json.Unmarshal([]byte(str), p)
	return
}

type periodTaskHandler struct {
	tk Worker
}

type Payload struct {
	Group   string `json:"group"`
	UID     string `json:"uid"`
	Payload string `json:"payload"`
}

type StreamPayload struct {
	TraceID         string         `json:"traceID,omitempty"`
	SpanID          string         `json:"spanID,omitempty"`
	UID             string         `json:"uid,omitempty"`
	Group           string         `json:"group,omitempty"`
	Payload         string         `json:"payload,omitempty"`
	Retention       int            `json:"retention,string,omitempty"`
	Replace         string         `json:"replace,omitempty"`
	MaxRetry        int            `json:"maxRetry,string,omitempty"`
	MaxArchivedTime int            `json:"maxArchivedTime,string,omitempty"`
	Timeout         int            `json:"timeout,string,omitempty"`
	In              *time.Duration `json:"in,string,omitempty"`
	Now             string         `json:"now,omitempty"`
}

type OncePayload struct {
	TraceID string `json:"traceID,omitempty"`
	SpanID  string `json:"spanID,omitempty"`
	Payload string `json:"payload,omitempty"`
}

func (p Payload) String() (str string) {
	bs, _ := json.Marshal(p)
	str = string(bs)
	return
}

func (p periodTaskHandler) ProcessTask(ctx context.Context, t *asynq.Task) (err error) {
	uid := uuid.NewString()
	payload := Payload{
		UID: t.ResultWriter().TaskID(),
	}
	var group string
	if strings.HasSuffix(t.Type(), ".once") {
		group = strings.TrimSuffix(t.Type(), ".once")
		var oncePayload OncePayload
		_ = json.Unmarshal(t.Payload(), &oncePayload)
		traceFromHex, _ := trace.TraceIDFromHex(oncePayload.TraceID)
		spanFromHex, _ := trace.SpanIDFromHex(oncePayload.SpanID)
		sc := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    traceFromHex,
			SpanID:     spanFromHex,
			TraceFlags: trace.FlagsSampled,
		})
		ctx = trace.ContextWithRemoteSpanContext(ctx, sc)
		payload.Payload = oncePayload.Payload
	} else {
		group = strings.TrimSuffix(t.Type(), ".cron")
		payload.Payload = string(t.Payload())
	}
	payload.Group = group
	tr := otel.Tracer("worker")
	ctx, span := tr.Start(ctx, "ProcessTask")
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()
	span.SetAttributes(
		attribute.String("uid", uid),
		attribute.String("group", group),
		attribute.String("payload", payload.Payload),
	)
	fields := log.Fields{
		"task": payload,
		"uuid": uid,
	}
	defer func() {
		if err != nil {
			log.
				WithContext(ctx).
				WithFields(fields).
				Warn("run task failed: %v", err)
			return
		}
		log.
			WithContext(ctx).
			WithFields(fields).
			Debug("run task success")
	}()
	if p.tk.ops.handler != nil {
		err = p.tk.ops.handler(ctx, payload)
	} else if p.tk.ops.handlerNeedWorker != nil {
		err = p.tk.ops.handlerNeedWorker(ctx, p.tk, payload)
	} else if p.tk.ops.callback != "" {
		err = p.httpCallback(ctx, payload)
	} else {
		log.
			WithContext(ctx).
			WithFields(fields).
			Info("no task handler")
	}
	// save processed count
	p.tk.processed(ctx, payload.UID)
	return
}

func (p periodTaskHandler) httpCallback(ctx context.Context, payload Payload) (err error) {
	client := &http.Client{}
	body := payload.String()
	var r *http.Request
	r, _ = http.NewRequestWithContext(ctx, http.MethodPost, p.tk.ops.callback, bytes.NewReader([]byte(body)))
	r.Header.Add("Content-Type", "application/json")
	var res *http.Response
	res, err = client.Do(r)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err = ErrHTTPCallbackInvalidStatusCode
	}
	return
}

// New is create a task worker, implemented by asynq: https://github.com/hibiken/asynq
func New(options ...func(*Options)) (tk *Worker) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	tk = &Worker{}
	if ops.redisURI == "" {
		tk.Error = errors.WithStack(ErrRedisNil)
		return
	}
	rs, err := asynq.ParseRedisURI(ops.redisURI)
	if err != nil {
		tk.Error = errors.WithStack(ErrRedisInvalid)
		return
	}
	// add group prefix to spilt difference group
	ops.redisPeriodKey = strings.Join([]string{ops.group, ops.redisPeriodKey}, ".")
	rds := rs.MakeRedisClient().(redis.UniversalClient)
	client := asynq.NewClient(rs)
	inspector := asynq.NewInspector(rs)
	// initialize redis lock
	locker := redislock.New(rds)
	// initialize basic fields first
	tk.ops = *ops
	tk.redis = rds
	tk.redisOpt = rs
	tk.locker = locker
	tk.client = client
	tk.inspector = inspector
	streamKey := strings.Join([]string{ops.redisPeriodKey, "waiting.stream"}, ".")
	tk.stream = stream.New(
		stream.WithRDS(rds),
		stream.WithKey(streamKey),
		stream.WithExpire(6*time.Hour),
	)
	var rpsInterval int64 = 30
	tk.streamLimiter = ratecounter.NewRateCounter(time.Duration(rpsInterval) * time.Second)
	// initialize scanner
	go func() {
		for {
			time.Sleep(ops.scanTaskInterval)
			tk.scan()
		}
	}()
	if tk.ops.clearArchived > 0 {
		// initialize clear archived
		go func() {
			for {
				time.Sleep(time.Duration(tk.ops.clearArchived) * time.Second)
				tk.clearArchived()
			}
		}()
	}
	go func() {
		for {
			// remove old data, no need execute
			tk.stream.Trim(context.Background(), int64(ops.streamMaxCount))

			// start consume
			tk.consumeOneWaiting()

			rps := int(tk.streamLimiter.Rate() / rpsInterval)

			if rps > ops.streamRPS {
				time.Sleep(10 * time.Second)
				continue
			}
			time.Sleep(3 * time.Second)
		}
	}()
	// initialize server after stream is set
	srv := asynq.NewServer(
		rs,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				ops.group: 10,
			},
			RetryDelayFunc:           ops.retryDelayFunc,
			DelayedTaskCheckInterval: ops.delayedTaskCheckInterval,
			Logger:                   myLogger{},
			LogLevel:                 levelToAsynq(ops.logLevel),
		},
	)
	go func() {
		var h periodTaskHandler
		h.tk = *tk
		if e := srv.Run(h); e != nil {
			log.WithError(err).Error("run task handler failed")
		}
	}()
	return
}

func (wk Worker) OnceWaiting(ctx context.Context, options ...func(*RunOptions)) (err error) {
	tr := otel.Tracer("worker")
	ctx, span := tr.Start(ctx, "OnceWaiting")
	var traceID, spanID string
	if s := trace.SpanContextFromContext(ctx); s.HasTraceID() {
		traceID = s.TraceID().String()
		spanID = s.SpanID().String()
	}
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()
	ops := getRunOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	span.SetAttributes(
		attribute.String("uid", ops.uid),
		attribute.String("group", ops.group),
		attribute.String("payload", ops.payload),
	)
	if ops.uid == "" {
		err = errors.WithStack(ErrUUIDNil)
		return
	}

	// only send to wait stream, not execute
	payload := StreamPayload{
		TraceID:         traceID,
		SpanID:          spanID,
		UID:             ops.uid,
		Group:           ops.group,
		Payload:         ops.payload,
		Retention:       ops.retention,
		Replace:         strconv.FormatBool(ops.replace),
		MaxRetry:        ops.maxRetry,
		MaxArchivedTime: ops.maxArchivedTime,
		Timeout:         ops.timeout,
		Now:             strconv.FormatBool(ops.now),
	}
	if ops.in != nil {
		payload.In = ops.in
	}
	err = wk.stream.Pub(ctx, payload)
	wk.streamLimiter.Incr(1)
	return
}

func (wk Worker) Once(ctx context.Context, options ...func(*RunOptions)) (err error) {
	tr := otel.Tracer("worker")
	ctx, span := tr.Start(ctx, "Once")
	var traceID, spanID string
	if s := trace.SpanContextFromContext(ctx); s.HasTraceID() {
		traceID = s.TraceID().String()
		spanID = s.SpanID().String()
	}
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	ops := getRunOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	span.SetAttributes(
		attribute.String("uid", ops.uid),
		attribute.String("group", ops.group),
		attribute.String("payload", ops.payload),
	)
	if ops.uid == "" {
		err = errors.WithStack(ErrUUIDNil)
		return
	}
	lock, err := wk.lock(ctx, ops.uid, *ops)
	if err != nil {
		return
	}
	defer func() {
		_ = lock.Release(ctx)
	}()
	payload, _ := json.Marshal(OncePayload{
		TraceID: traceID,
		SpanID:  spanID,
		Payload: ops.payload,
	})
	t := asynq.NewTask(strings.Join([]string{ops.group, "once"}, "."), payload, asynq.TaskID(ops.uid))
	taskOpts := []asynq.Option{
		asynq.Queue(wk.ops.group),
		asynq.MaxRetry(wk.ops.maxRetry),
		asynq.Timeout(time.Duration(ops.timeout) * time.Second),
	}
	if ops.maxRetry > 0 {
		taskOpts = append(taskOpts, asynq.MaxRetry(ops.maxRetry))
	}
	if ops.retention > 0 {
		taskOpts = append(taskOpts, asynq.Retention(time.Duration(ops.retention)*time.Second))
	} else {
		taskOpts = append(taskOpts, asynq.Retention(time.Duration(wk.ops.retention)*time.Second))
	}
	if ops.in != nil {
		taskOpts = append(taskOpts, asynq.ProcessIn(*ops.in))
	} else if ops.at != nil {
		taskOpts = append(taskOpts, asynq.ProcessAt(*ops.at))
	} else if ops.now {
		taskOpts = append(taskOpts, asynq.ProcessIn(time.Millisecond))
	}
	info, err := wk.inspector.GetTaskInfo(wk.ops.group, ops.uid)
	if err != nil && !strings.Contains(err.Error(), asynq.ErrQueueNotFound.Error()) && !strings.Contains(err.Error(), asynq.ErrTaskNotFound.Error()) {
		// other error
		return
	} else if err != nil && (strings.Contains(err.Error(), asynq.ErrQueueNotFound.Error()) || strings.Contains(err.Error(), asynq.ErrTaskNotFound.Error())) {
		// no queue or no task
		_, err = wk.client.Enqueue(t, taskOpts...)
		return
	}
	// task exists
	if info.State == asynq.TaskStateActive {
		return
	}
	if ops.replace {
		// remove old one if replace = true
		err = wk.Remove(context.Background(), ops.uid)
		if err != nil {
			return
		}
		_, err = wk.client.Enqueue(t, taskOpts...)
	}
	return
}

func (wk Worker) Cron(ctx context.Context, options ...func(*RunOptions)) (err error) {
	ops := getRunOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.uid == "" {
		err = errors.WithStack(ErrUUIDNil)
		return
	}

	// Get expressions from options
	exprs := ops.exprs
	if len(exprs) == 0 {
		err = errors.WithStack(ErrExprInvalid)
		return
	}

	// Validate expressions for duplicates/conflicts
	if err = validateExprs(exprs); err != nil {
		return
	}

	var next int64
	next, _, err = getNextMulti(exprs, 0)
	if err != nil {
		err = errors.WithStack(ErrExprInvalid)
		return
	}
	lock, err := wk.lock(ctx, ops.uid, *ops)
	if err != nil {
		return
	}
	defer func() {
		_ = lock.Release(ctx)
	}()
	// check if the task has been dynamically modified
	// if OriginalExprs is set and matches the configured exprs, skip overwriting
	existingTask, e := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, ops.uid).Result()
	if e == nil {
		var existing periodTask
		existing.FromString(existingTask)

		// Check if task was dynamically modified
		if len(existing.OriginalExprs) > 0 {
			// Compare arrays
			if exprsEqual(existing.OriginalExprs, exprs) && !exprsEqual(existing.Exprs, exprs) {
				log.WithContext(ctx).WithFields(log.Fields{
					"uid":           ops.uid,
					"originalExprs": existing.OriginalExprs,
					"currentExprs":  existing.Exprs,
					"configExprs":   exprs,
				}).Info("skip overwriting cron task: task has been dynamically modified")
				return
			}
		}
	}

	t := periodTask{
		Exprs:         exprs,
		OriginalExprs: nil, // empty on creation, set by UpdateCronExpr on first modification
		Group:         strings.Join([]string{ops.group, "cron"}, "."),
		UID:           ops.uid,
		Payload:       ops.payload,
		Next:          next,
		MaxRetry:      ops.maxRetry,
		Timeout:       ops.timeout,
	}

	// remove old task
	_ = wk.Remove(context.Background(), t.UID)
	_, err = wk.redis.HSet(ctx, wk.ops.redisPeriodKey, ops.uid, t.String()).Result()
	if err != nil {
		err = errors.WithStack(ErrSaveCron)
		return
	}
	return
}

func (wk Worker) Remove(ctx context.Context, uid string) (err error) {
	wk.redis.HDel(ctx, wk.ops.redisPeriodKey, uid)
	e1 := wk.inspector.CancelProcessing(uid)
	if e1 != nil {
		log.WithContext(ctx).Warn("cancel processing failed: %v", e1)
	}
	e2 := wk.inspector.DeleteTask(wk.ops.group, uid)
	if e2 != nil {
		log.WithContext(ctx).Warn("delete task failed: %v", e2)
	}
	return
}

// UpdateCronExpr dynamically updates the cron expression(s) for an existing cron task.
// Supports both single and multiple expressions.
// The original expressions are preserved for later restoration.
//
// uid: unique identifier of the cron task
// newExpr: new cron expression(s) to apply (can pass one or more expressions)
func (wk Worker) UpdateCronExpr(ctx context.Context, uid string, newExpr ...string) (err error) {
	if uid == "" {
		err = errors.WithStack(ErrUUIDNil)
		return
	}

	if len(newExpr) == 0 {
		err = errors.WithStack(ErrExprInvalid)
		return
	}

	// Validate the new expressions
	if err = validateExprs(newExpr); err != nil {
		return
	}

	// Calculate next execution time for new expressions
	newNext, _, newInterval, err := getNextFromExprs(newExpr, 0)
	if err != nil {
		err = errors.WithStack(ErrExprInvalid)
		return
	}

	// acquire lock to prevent concurrent modifications
	lock, err := wk.lock(ctx, uid, RunOptions{
		lockerTTL:           wk.ops.lockerTTL,
		lockerRetryCount:    wk.ops.lockerRetryCount,
		lockerRetryInterval: wk.ops.lockerRetryInterval,
	})
	if err != nil {
		return
	}
	defer func() {
		_ = lock.Release(ctx)
	}()

	// retrieve existing task from redis
	t, err := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, uid).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = errors.WithStack(ErrCronTaskNotFound)
		}
		return
	}

	var task periodTask
	task.FromString(t)

	// skip if newExpr is the same as current running exprs
	if exprsEqual(task.Exprs, newExpr) {
		log.WithContext(ctx).WithFields(log.Fields{
			"uid":   uid,
			"exprs": newExpr,
		}).Info("skip updating cron task: exprs are already the same")
		return
	}

	// preserve original expressions on first modification
	if len(task.OriginalExprs) == 0 {
		task.OriginalExprs = task.Exprs
	}

	// calculate old expression interval
	oldNext, _, oldInterval, _ := getNextFromExprs(task.Exprs, 0)

	// determine next execution time based on interval comparison
	var next int64
	if newInterval > oldInterval {
		// delay task: newExpr interval > oldExpr interval
		if newNext <= oldNext {
			// newExpr.next is before or equal to oldExpr.next, use newExpr.next.next
			next = newNext + newInterval
		} else {
			// newExpr.next is after oldExpr.next, use newExpr.next
			next = newNext
		}
	} else {
		// advance task or same interval: use newExpr.next
		next = newNext
	}

	// apply the new expressions
	task.Exprs = newExpr
	task.Next = next

	// remove queued task to allow rescheduling with new expression
	_ = wk.inspector.DeleteTask(wk.ops.group, uid)

	// persist changes to redis
	_, err = wk.redis.HSet(ctx, wk.ops.redisPeriodKey, uid, task.String()).Result()
	if err != nil {
		err = errors.WithStack(ErrSaveCron)
		return
	}
	return
}

// RestoreCronExpr restores the cron expression to its original value.
// This reverses any changes made by UpdateCronExpr.
// uid: unique identifier of the cron task
func (wk Worker) RestoreCronExpr(ctx context.Context, uid string) (err error) {
	if uid == "" {
		err = errors.WithStack(ErrUUIDNil)
		return
	}
	// acquire lock to prevent concurrent modifications
	lock, err := wk.lock(ctx, uid, RunOptions{
		lockerTTL:           wk.ops.lockerTTL,
		lockerRetryCount:    wk.ops.lockerRetryCount,
		lockerRetryInterval: wk.ops.lockerRetryInterval,
	})
	if err != nil {
		return
	}
	defer func() {
		_ = lock.Release(ctx)
	}()
	// retrieve existing task from redis
	t, err := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, uid).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = errors.WithStack(ErrCronTaskNotFound)
		}
		return
	}
	var task periodTask
	task.FromString(t)

	// if no original expressions exist, nothing to restore
	if len(task.OriginalExprs) == 0 {
		return
	}

	// validate and calculate next execution time for original expressions
	next, _, err := getNextMulti(task.OriginalExprs, 0)
	if err != nil {
		err = errors.WithStack(ErrExprInvalid)
		return
	}

	// restore to original expressions
	task.Exprs = task.OriginalExprs
	task.OriginalExprs = nil // clear to indicate restored state
	task.Next = next

	// remove queued task to allow rescheduling with restored expression
	_ = wk.inspector.DeleteTask(wk.ops.group, uid)

	// persist changes to redis
	_, err = wk.redis.HSet(ctx, wk.ops.redisPeriodKey, uid, task.String()).Result()
	if err != nil {
		err = errors.WithStack(ErrSaveCron)
		return
	}
	return
}


func (wk Worker) processed(ctx context.Context, uid string) {
	lock, err := wk.lock(ctx, strings.Join([]string{"processed", uid}, "."), RunOptions{
		lockerTTL:           wk.ops.lockerTTL,
		lockerRetryCount:    wk.ops.lockerRetryCount,
		lockerRetryInterval: wk.ops.lockerRetryInterval,
	})
	if err != nil {
		return
	}
	defer func() {
		_ = lock.Release(context.Background())
	}()
	t, e := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, uid).Result()
	if e == nil || !errors.Is(e, redis.Nil) {
		var item periodTask
		item.FromString(t)
		item.Processed++
		wk.redis.HSet(ctx, wk.ops.redisPeriodKey, uid, item.String())
	}
	return
}

func (wk Worker) scan() {
	ctx := context.Background()
	tr := otel.Tracer("worker")
	ctx, span := tr.Start(ctx, "scan")
	defer span.End()
	lock, err := wk.lock(ctx, "__internal_worker_scan", RunOptions{
		lockerTTL:           wk.ops.lockerTTL,
		lockerRetryCount:    wk.ops.lockerRetryCount,
		lockerRetryInterval: wk.ops.lockerRetryInterval,
	})
	if err != nil {
		return
	}
	defer func() {
		_ = lock.Release(context.Background())
	}()
	m, _ := wk.redis.HGetAll(ctx, wk.ops.redisPeriodKey).Result()
	p := wk.redis.Pipeline()
	ops := wk.ops
	for _, v := range m {
		var item periodTask
		item.FromString(v)
		if wk.hasTask(item.UID) {
			continue
		}

		if len(item.Exprs) == 0 {
			continue
		}

		// Calculate next execution time using multiple expressions if available
		next, diff, _ := getNextMulti(item.Exprs, item.Next)

		t := asynq.NewTask(item.Group, []byte(item.Payload), asynq.TaskID(item.UID))
		taskOpts := []asynq.Option{
			asynq.Queue(ops.group),
			asynq.MaxRetry(ops.maxRetry),
			asynq.Timeout(time.Duration(item.Timeout) * time.Second),
		}
		if item.MaxRetry > 0 {
			taskOpts = append(taskOpts, asynq.MaxRetry(item.MaxRetry))
		}
		if diff > 10 {
			retention := diff / 3
			if diff > 600 {
				// max retention 10min
				retention = 600
			}
			// set retention avoid repeat in short time
			taskOpts = append(taskOpts, asynq.Retention(time.Duration(retention)*time.Second))
		}
		taskOpts = append(taskOpts, asynq.ProcessAt(time.Unix(item.Next, 0)))
		_, err = wk.client.Enqueue(t, taskOpts...)
		// enqueue success, update next
		if err == nil {
			item.Next = next
			p.HSet(ctx, wk.ops.redisPeriodKey, item.UID, item.String())
		}
	}
	// batch save to cache
	p.Exec(ctx)
	return
}

func (wk Worker) hasTask(id string) bool {
	task, _ := wk.inspector.GetTaskInfo(wk.ops.group, id)
	if task != nil {
		return true
	}
	return false
}

func (wk Worker) clearArchived() {
	list, err := wk.inspector.ListArchivedTasks(wk.ops.group, asynq.Page(1), asynq.PageSize(100))
	if err != nil {
		return
	}
	ctx := context.Background()
	for _, item := range list {
		last := carbon.CreateFromStdTime(item.LastFailedAt)
		if !last.IsZero() && item.Retried < item.MaxRetry {
			continue
		}
		uid := item.ID
		var archivedTime int
		if strings.HasSuffix(item.Type, ".cron") {
			// cron task
			t, e := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, uid).Result()
			if e == nil || !errors.Is(e, redis.Nil) {
				var task periodTask
				task.FromString(t)

				if len(task.Exprs) > 0 {
					_, diff, _ := getNextMulti(task.Exprs, task.Next)
					// default archived 1/2 task interval
					archivedTime = int((diff) / 2)
					if task.MaxArchivedTime > 0 {
						archivedTime = task.MaxArchivedTime
					}
				}
			}
		} else {
			// once task default archived 5 minute
			archivedTime = 300
			if wk.ops.maxArchivedTime > 0 {
				archivedTime = wk.ops.maxArchivedTime
			}
		}
		if carbon.Now().Gt(last.AddSeconds(archivedTime)) {
			_ = wk.Remove(ctx, uid)
		}
	}
}

func (wk Worker) consumeOneWaiting() {
	var err error
	ctx := context.Background()
	tr := otel.Tracer("worker")
	ctx, span := tr.Start(ctx, "consumeOneWaiting")
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()
	msgs, err := wk.stream.ReadBatch(ctx, int64(wk.ops.streamMaxCount))
	if err != nil {
		return
	}
	var lastID string
	for _, msg := range msgs {
		var data StreamPayload
		str, _ := json.Marshal(msg.Values)
		_ = json.Unmarshal(str, &data)
		traceFromHex, _ := trace.TraceIDFromHex(data.TraceID)
		spanFromHex, _ := trace.SpanIDFromHex(data.SpanID)
		sc := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    traceFromHex,
			SpanID:     spanFromHex,
			TraceFlags: trace.FlagsSampled,
		})
		onceCtx := trace.ContextWithRemoteSpanContext(context.Background(), sc)

		options := []func(*RunOptions){
			WithRunUUID(data.UID),
			WithRunGroup(data.Group),
			WithRunPayload(data.Payload),
			WithRunRetention(data.Retention),
			WithRunMaxRetry(data.MaxRetry),
			WithRunMaxArchivedTime(data.MaxArchivedTime),
			WithRunTimeout(data.Timeout),
		}
		if data.Replace == "true" {
			options = append(options, WithRunReplace(true))
		}
		if data.Now == "true" {
			options = append(options, WithRunNow(true))
		}
		if data.In != nil {
			options = append(options, WithRunIn(*data.In))
		}
		// add once task
		err = wk.Once(onceCtx, options...)
		lastID = msg.ID
	}
	if lastID == "" {
		return
	}
	wk.stream.TrimLteMinID(context.Background(), lastID)
}

func (wk Worker) lock(ctx context.Context, prefix string, ops RunOptions) (*redislock.Lock, error) {
	tr := otel.Tracer("worker")
	ctx, span := tr.Start(ctx, "lock")
	defer span.End()
	key := strings.Join([]string{wk.ops.redisPeriodKey, prefix, "lock"}, ".")
	lock, err := wk.locker.Obtain(
		ctx,
		key,
		ops.lockerTTL,
		&redislock.Options{
			RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(ops.lockerRetryInterval), ops.lockerRetryCount),
		},
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	return lock, err
}

func getNext(expr string, timestamp int64) (end, diff int64, err error) {
	var e *cronexpr.Expression
	e, err = cronexpr.Parse(expr)
	if err != nil {
		return
	}
	now := carbon.Now()
	nowTimestamp := now.Timestamp()
	t := now.StdTime()
	start := nowTimestamp
	if timestamp > 0 {
		t = carbon.CreateFromTimestamp(timestamp).StdTime()
		start = timestamp
	}
	end = e.Next(t).Unix()
	// time has expired
	if end < nowTimestamp {
		end = e.Next(now.StdTime()).Unix()
		start = nowTimestamp
	}
	// calc diff
	diff = end - start
	return
}

// getNextMulti is a wrapper that handles both single and multiple expressions
// For backward compatibility with existing code
func getNextMulti(exprs []string, timestamp int64) (end, diff int64, err error) {
	if len(exprs) == 0 {
		err = errors.WithStack(ErrExprInvalid)
		return
	}

	if len(exprs) == 1 {
		// Single expression - use original logic
		return getNext(exprs[0], timestamp)
	}

	// Multiple expressions - find nearest next time
	var matchedExpr string
	var interval int64
	end, matchedExpr, interval, err = getNextFromExprs(exprs, timestamp)
	if err != nil {
		return
	}

	now := carbon.Now()
	nowTimestamp := now.Timestamp()
	start := nowTimestamp
	if timestamp > 0 {
		start = timestamp
	}

	// Calculate diff
	diff = end - start
	if diff < 0 {
		diff = interval
	}

	_ = matchedExpr // used for debugging if needed
	return
}

// exprsEqual compares two string slices for equality
func exprsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// validateExprs validates that cron expressions don't have duplicates or conflicts
// Returns error if validation fails
func validateExprs(exprs []string) error {
	if len(exprs) == 0 {
		return errors.WithStack(ErrExprInvalid)
	}

	// Parse all expressions first to validate syntax
	parsedExprs := make([]*cronexpr.Expression, 0, len(exprs))
	for _, expr := range exprs {
		e, err := cronexpr.Parse(expr)
		if err != nil {
			return errors.WithStack(ErrExprInvalid)
		}
		parsedExprs = append(parsedExprs, e)
	}

	// Check for duplicate next execution times
	// We'll check the next 100 execution times for each expression
	now := carbon.Now().StdTime()
	executionTimes := make(map[int64]string) // timestamp -> expr

	for i, e := range parsedExprs {
		t := now
		for j := 0; j < 100; j++ {
			next := e.Next(t)
			nextUnix := next.Unix()

			if existingExpr, exists := executionTimes[nextUnix]; exists {
				// Found duplicate execution time
				return errors.Errorf("duplicate execution time detected: expr[%d]='%s' and expr='%s' both execute at %s",
					i, exprs[i], existingExpr, next.Format("2006-01-02 15:04:05"))
			}
			executionTimes[nextUnix] = exprs[i]
			t = next
		}
	}

	return nil
}

// getNextFromExprs finds the nearest next execution time from multiple cron expressions
// Returns the next execution time, the expression that produces it, and the interval
func getNextFromExprs(exprs []string, timestamp int64) (next int64, matchedExpr string, interval int64, err error) {
	if len(exprs) == 0 {
		err = errors.WithStack(ErrExprInvalid)
		return
	}

	now := carbon.Now()
	nowTimestamp := now.Timestamp()
	baseTime := now.StdTime()
	start := nowTimestamp

	if timestamp > 0 {
		baseTime = carbon.CreateFromTimestamp(timestamp).StdTime()
		start = timestamp
	}

	var minNext int64 = 0
	var minExpr string
	var minInterval int64 = 0

	// Iterate through all expressions to find the nearest next time
	for _, expr := range exprs {
		e, parseErr := cronexpr.Parse(expr)
		if parseErr != nil {
			err = parseErr
			return
		}

		nextTime := e.Next(baseTime)
		nextUnix := nextTime.Unix()

		// If time has expired, use current time
		if nextUnix < nowTimestamp {
			nextTime = e.Next(now.StdTime())
			nextUnix = nextTime.Unix()
		}

		// Calculate interval for this expression
		secondNext := e.Next(nextTime)
		exprInterval := secondNext.Unix() - nextUnix

		// Find the minimum (earliest) next time
		if minNext == 0 || nextUnix < minNext {
			minNext = nextUnix
			minExpr = expr
			minInterval = exprInterval
		}
	}

	next = minNext
	matchedExpr = minExpr
	interval = minInterval

	// Recalculate diff based on start time
	if next < start {
		next = minNext
	}

	return
}
