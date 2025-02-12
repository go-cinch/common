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
	Expr            string `json:"expr"` // cron expr github.com/gorhill/cronexpr
	Group           string `json:"group"`
	UID             string `json:"uid"`
	Payload         string `json:"payload"`
	Next            int64  `json:"next"`      // next schedule unix timestamp
	Processed       int64  `json:"processed"` // run times
	MaxRetry        int    `json:"maxRetry"`
	MaxArchivedTime int    `json:"maxArchivedTime"`
	Timeout         int    `json:"timeout"`
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
	// initialize server
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
	var next int64
	next, _, err = getNext(ops.expr, 0)
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
	t := periodTask{
		Expr:     ops.expr,
		Group:    strings.Join([]string{ops.group, "cron"}, "."),
		UID:      ops.uid,
		Payload:  ops.payload,
		Next:     next,
		MaxRetry: ops.maxRetry,
		Timeout:  ops.timeout,
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
	err = wk.inspector.DeleteTask(wk.ops.group, uid)
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
		next, diff, _ := getNext(item.Expr, item.Next)
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
				_, diff, _ := getNext(task.Expr, task.Next)
				// default archived 1/2 task interval
				archivedTime = int((diff) / 2)
				if task.MaxArchivedTime > 0 {
					archivedTime = task.MaxArchivedTime
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
