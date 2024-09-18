package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-cinch/common/log"
	"github.com/go-cinch/common/nx"
	"github.com/golang-module/carbon/v2"
	"github.com/google/uuid"
	"github.com/gorhill/cronexpr"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

type Worker struct {
	ops       Options
	redis     redis.UniversalClient
	redisOpt  asynq.RedisConnOpt
	lock      *nx.Nx
	client    *asynq.Client
	inspector *asynq.Inspector
	Error     error
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

func (p Payload) String() (str string) {
	bs, _ := json.Marshal(p)
	str = string(bs)
	return
}

func (p periodTaskHandler) ProcessTask(ctx context.Context, t *asynq.Task) (err error) {
	tr := otel.Tracer("worker")
	ctx, span := tr.Start(ctx, "ProcessTask")
	defer span.End()
	uid := uuid.NewString()
	group := strings.TrimSuffix(strings.TrimSuffix(t.Type(), ".once"), ".cron")
	payload := Payload{
		Group:   group,
		UID:     t.ResultWriter().TaskID(),
		Payload: string(t.Payload()),
	}
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
	rd := rs.MakeRedisClient().(redis.UniversalClient)
	client := asynq.NewClient(rs)
	inspector := asynq.NewInspector(rs)
	// initialize redis lock
	nxLock := nx.New(
		nx.WithRedis(rd),
		nx.WithExpire(10),
		nx.WithKey(strings.Join([]string{ops.redisPeriodKey, "lock"}, ".")),
	)
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
	tk.redis = rd
	tk.redisOpt = rs
	tk.lock = nxLock
	tk.client = client
	tk.inspector = inspector
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
	return
}

func (wk Worker) Once(options ...func(*RunOptions)) (err error) {
	ops := getRunOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.uid == "" {
		err = errors.WithStack(ErrUUIDNil)
		return
	}
	err = wk.lock.MustLock()
	if err != nil {
		return
	}
	defer wk.lock.Unlock()
	t := asynq.NewTask(strings.Join([]string{ops.group, "once"}, "."), []byte(ops.payload), asynq.TaskID(ops.uid))
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
	_, err = wk.client.Enqueue(t, taskOpts...)
	if ops.replace && errors.Is(err, asynq.ErrTaskIDConflict) {
		// remove old one if replace = true
		ctx := wk.getDefaultTimeoutCtx()
		if ops.ctx != nil {
			ctx = ops.ctx
		}
		err = wk.Remove(ctx, ops.uid)
		if err != nil {
			return
		}
		_, err = wk.client.Enqueue(t, taskOpts...)
	}
	return
}

func (wk Worker) Cron(options ...func(*RunOptions)) (err error) {
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
	err = wk.lock.MustLock()
	if err != nil {
		return
	}
	defer wk.lock.Unlock()
	t := periodTask{
		Expr:     ops.expr,
		Group:    strings.Join([]string{ops.group, "cron"}, "."),
		UID:      ops.uid,
		Payload:  ops.payload,
		Next:     next,
		MaxRetry: ops.maxRetry,
		Timeout:  ops.timeout,
	}
	ctx := wk.getDefaultTimeoutCtx()
	// remove old task
	wk.Remove(ctx, t.UID)
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
	err := wk.lock.MustLock(ctx)
	if err != nil {
		return
	}
	defer wk.lock.Unlock(ctx)
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
	ctx := wk.getDefaultTimeoutCtx()
	ok := wk.lock.Lock()
	if !ok {
		return
	}
	defer wk.lock.Unlock()
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
		_, err := wk.client.Enqueue(t, taskOpts...)
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
	ctx := wk.getDefaultTimeoutCtx()
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
			wk.inspector.DeleteTask(wk.ops.group, uid)
		}
	}
}

func (wk Worker) getDefaultTimeoutCtx() context.Context {
	c, _ := context.WithTimeout(context.Background(), time.Duration(wk.ops.timeout)*time.Second)
	return c
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
