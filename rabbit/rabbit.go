package rabbit

import (
	"github.com/go-cinch/common/log"
	"github.com/google/uuid"
	"github.com/houseofcat/turbocookedrabbit/v2/pkg/tcr"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"sync/atomic"
	"time"
)

type Rabbit struct {
	ops        Options
	pool       *tcr.ConnectionPool
	poolConfig *tcr.PoolConfig
	healthHost *tcr.ConnectionHost
	lost       int32
	Error      error
}

type Exchange struct {
	ops   ExchangeOptions
	rb    *Rabbit
	Error error
}

type Queue struct {
	ops   QueueOptions
	ex    *Exchange
	Error error
}

func New(dsn string, options ...func(*Options)) (rb *Rabbit) {
	rb = &Rabbit{}
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	rb.ops = *ops
	name := ops.name
	if name == "" {
		name = uuid.NewString()[:8]
	}
	rb.poolConfig = &tcr.PoolConfig{
		ApplicationName:      name,
		URI:                  dsn,
		Heartbeat:            uint32(ops.heartbeat),
		ConnectionTimeout:    uint32(ops.timeout),
		SleepOnErrorInterval: uint32(ops.healthCheckInterval),
		MaxConnectionCount:   uint64(ops.maxConnection),
		MaxCacheChannelCount: uint64(ops.maxChannel),
	}
	healthPoolConfig := &tcr.PoolConfig{
		ApplicationName:      name + "-hc",
		URI:                  dsn,
		Heartbeat:            uint32(ops.heartbeat),
		ConnectionTimeout:    uint32(ops.timeout),
		SleepOnErrorInterval: uint32(ops.healthCheckInterval),
		MaxConnectionCount:   1,
		MaxCacheChannelCount: 1,
	}
	pool, err := tcr.NewConnectionPoolWithErrorHandler(
		rb.poolConfig,
		func(err error) {
			log.WithContext(rb.ops.ctx).WithError(err).Error("rabbit pool err")
		},
	)

	if err != nil {
		rb.Error = err
		return
	}
	healthPool, _ := tcr.NewConnectionPool(healthPoolConfig)
	rb.healthHost, err = healthPool.GetConnection()
	if err != nil {
		rb.Error = err
		return
	}
	rb.pool = pool
	go rb.healthCheck()
	return
}

func (rb *Rabbit) healthCheck() {
	// InfiniteLoop: Stay here till we reconnect.
	for {
		ok := rb.healthHost.Connect()
		if ok {
			atomic.CompareAndSwapInt32(&rb.lost, 1, 0)
		} else {
			atomic.CompareAndSwapInt32(&rb.lost, 0, 1)
		}
		time.Sleep(time.Duration(rb.ops.healthCheckInterval) * time.Millisecond)
	}
}

func (rb *Rabbit) Ping() (err error) {
	if rb.lost == 1 {
		err = errors.Errorf("connection maybe lost")
		return
	}
	return
}

// Exchange bind a exchange
func (rb *Rabbit) Exchange(options ...func(*ExchangeOptions)) *Exchange {
	ex := rb.beforeExchange(options...)
	if ex.Error != nil {
		return ex
	}
	// the exchange will be declared
	if ex.ops.declare {
		err := ex.declare()
		if err != nil {
			ex.Error = errors.WithStack(err)
			return ex
		}
	}
	return ex
}

// before bind exchange
func (rb *Rabbit) beforeExchange(options ...func(*ExchangeOptions)) *Exchange {
	var ex Exchange
	if rb.Error != nil {
		ex.Error = rb.Error
		return &ex
	}
	ops := getExchangeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.name == "" {
		ex.Error = errors.Errorf("exchange name is empty")
		return &ex
	}
	switch ops.kind {
	case amqp.ExchangeDirect:
	case amqp.ExchangeFanout:
	case amqp.ExchangeTopic:
	case amqp.ExchangeHeaders:
	default:
		ex.Error = errors.Errorf("invalid exchange kind: %s", ops.kind)
		return &ex
	}
	prefix := ""
	if ops.namePrefix != "" {
		prefix = ops.namePrefix
	}
	ops.name = prefix + ops.name
	ex.ops = *ops
	ex.rb = rb
	return &ex
}

// Queue bind a queue
func (ex *Exchange) Queue(options ...func(*QueueOptions)) *Queue {
	qu := ex.beforeQueue(options...)
	if qu.Error != nil {
		return qu
	}
	if qu.ops.declare {
		err := qu.declare()
		if err != nil {
			qu.Error = errors.WithStack(err)
			return qu
		}
	}
	if qu.ops.bind {
		err := qu.bind()
		if err != nil {
			qu.Error = errors.WithStack(err)
			return qu
		}
	}
	return qu
}

// QueueWithDeadLetter bind a dead letter queue
func (ex *Exchange) QueueWithDeadLetter(options ...func(*QueueOptions)) *Queue {
	ops := getQueueOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	args := make(amqp.Table)
	if ops.args != nil {
		args = ops.args
	}

	if ops.deadLetterName == "" {
		var qu Queue
		qu.Error = errors.Errorf("dead letter name is empty")
		return &qu
	}
	args["x-dead-letter-exchange"] = ops.deadLetterName
	if ops.deadLetterKey != "" {
		args["x-dead-letter-routing-key"] = ops.deadLetterKey
	}
	ops.args = args
	return ex.Queue(func(options *QueueOptions) {
		*options = *ops
	})
}

// before bind queue
func (ex *Exchange) beforeQueue(options ...func(*QueueOptions)) *Queue {
	var qu Queue
	if ex.Error != nil {
		qu.Error = ex.Error
		return &qu
	}
	ops := getQueueOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	args := make(amqp.Table)
	if ops.args != nil {
		args = ops.args
	}
	if ops.messageTTL > 0 {
		args["x-message-ttl"] = ops.messageTTL
		ops.args = args
	}
	if ops.name == "" {
		qu.Error = errors.Errorf("queue name is empty")
		return &qu
	}
	prefix := ""
	if ops.namePrefix != "" {
		prefix = ops.namePrefix
	}
	ops.name = prefix + ops.name
	qu.ops = *ops
	qu.ex = ex
	return &qu
}

// declare exchange
func (ex *Exchange) declare() (err error) {
	ch := ex.rb.pool.GetChannelFromPool()
	defer func() {
		ex.rb.pool.ReturnChannel(ch, false)
	}()
	if err = ch.Channel.ExchangeDeclare(
		ex.ops.name,
		ex.ops.kind,
		ex.ops.durable,
		ex.ops.autoDelete,
		ex.ops.internal,
		ex.ops.noWait,
		ex.ops.args,
	); err != nil {
		err = errors.Wrapf(err, "failed declare exchange %s(%s)", ex.ops.name, ex.ops.kind)
		return
	}
	return
}

// declare queue
func (qu *Queue) declare() (err error) {
	ch := qu.ex.rb.pool.GetChannelFromPool()
	defer func() {
		qu.ex.rb.pool.ReturnChannel(ch, false)
	}()
	if _, err = ch.Channel.QueueDeclare(
		qu.ops.name,
		qu.ops.durable,
		qu.ops.autoDelete,
		qu.ops.exclusive,
		qu.ops.noWait,
		qu.ops.args,
	); err != nil {
		err = errors.Wrapf(err, "failed to declare %s", qu.ops.name)
		return
	}
	return
}

// bind queue
func (qu *Queue) bind() (err error) {
	ch := qu.ex.rb.pool.GetChannelFromPool()
	defer func() {
		qu.ex.rb.pool.ReturnChannel(ch, false)
	}()
	for _, key := range qu.ops.routeKeys {
		if err = ch.Channel.QueueBind(
			qu.ops.name,
			key,
			qu.ex.ops.name,
			qu.ops.noWait,
			qu.ops.args,
		); err != nil {
			err = errors.Wrapf(err, "failed to declare bind queue, queue: %s, key: %s, exchange: %s", qu.ops.name, key, qu.ex.ops.name)
			return
		}
	}
	return
}
