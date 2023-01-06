package rabbit

import (
	"context"
	"github.com/go-cinch/common/log"
	"github.com/houseofcat/turbocookedrabbit/v2/pkg/tcr"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"sync/atomic"
	"time"
)

type Consume struct {
	ops      ConsumeOptions
	q        string
	consumer *tcr.Consumer
	Error    error
}

func (qu *Queue) Consume(handler func(context.Context, string, amqp.Delivery) bool, options ...func(*ConsumeOptions)) (err error) {
	if handler == nil {
		err = errors.Errorf("handler is nil")
		return
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		err = errors.WithStack(co.Error)
		return
	}
	co.consumer.StartConsumingWithAction(func(msg *tcr.ReceivedMessage) {
		d := msg.Delivery
		a := d.Acknowledger
		tag := d.DeliveryTag
		ok := handler(context.Background(), co.q, d)
		if co.ops.autoAck {
			return
		}
		if ok {
			e := a.Ack(tag, false)
			if e != nil {
				log.WithError(e).Error("consume ack failed")
			}
			return
		}
		e := a.Nack(tag, false, co.ops.nackRequeue)
		if e != nil {
			log.WithError(e).Error("consume nack failed")
		}
	})
	return
}

func (qu *Queue) ConsumeOne(size int, handler func(c context.Context, q string, d amqp.Delivery) bool, options ...func(*ConsumeOptions)) (err error) {
	if size < 1 {
		err = errors.Errorf("minimum size is 1")
		return
	}
	if handler == nil {
		err = errors.Errorf("handler is nil")
		return
	}

	if qu.ex == nil {
		err = errors.Errorf("ex is nil")
		return
	}

	if atomic.LoadInt32(&qu.ex.rb.lost) == 1 {
		err = errors.Errorf("connection maybe lost")
		return
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		err = errors.WithStack(co.Error)
		return
	}
	ds := qu.getBatch(qu.ops.name, size, co.ops.autoAck)
	ctx := context.Background()
	if co.ops.oneCtx != nil {
		ctx = co.ops.oneCtx
	}
	for i, d := range ds {
		a := d.Acknowledger
		tag := d.DeliveryTag
		ctx = context.WithValue(ctx, "index", i)
		ok := handler(ctx, co.q, d)
		if co.ops.autoAck {
			return
		}
		if ok {
			e := a.Ack(tag, false)
			if e != nil {
				log.WithContext(ctx).WithError(e).Error("consume one ack failed")
			}
			return
		}
		// get retry count
		var retryCount int32
		if v, o := d.Headers["x-retry-count"].(int32); o {
			retryCount = v + 1
		} else {
			retryCount = 1
		}
		// need to retry, requeue with set custom header
		if co.ops.nackRetry {
			if retryCount < co.ops.nackMaxRetryCount {
				d.Headers["x-retry-count"] = retryCount
				err = qu.ex.PublishByte(
					d.Body,
					WithPublishHeaders(d.Headers),
					WithPublishRouteKey(d.RoutingKey),
				)
				if err != nil {
					log.WithContext(ctx).WithError(err).Error("consume one republish failed")
				}
			} else {
				log.WithContext(ctx).Warn("maximum retry %d exceeded, discard data", co.ops.nackMaxRetryCount)
			}
		}
		e := a.Nack(tag, false, co.ops.nackRequeue)
		if e != nil {
			log.WithContext(ctx).WithError(e).Error("consume one nack failed")
		}
	}
	return
}

func (qu *Queue) beforeConsume(options ...func(*ConsumeOptions)) *Consume {
	var co Consume
	if qu.Error != nil {
		co.Error = qu.Error
		return &co
	}
	ops := getConsumeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	co.ops = *ops
	co.q = qu.ops.name
	co.consumer = tcr.NewConsumerFromConfig(
		&tcr.ConsumerConfig{
			Enabled:              true,
			QueueName:            qu.ops.name,
			ConsumerName:         co.ops.consumer,
			AutoAck:              co.ops.autoAck,
			Exclusive:            co.ops.exclusive,
			NoWait:               co.ops.noWait,
			Args:                 co.ops.args,
			QosCountOverride:     co.ops.qosPrefetchCount,
			SleepOnIdleInterval:  uint32(qu.ex.rb.ops.healthCheckInterval),
			SleepOnErrorInterval: uint32(qu.ex.rb.ops.healthCheckInterval),
		},
		qu.ex.rb.pool,
	)
	return &co
}

func (qu *Queue) getBatch(queueName string, batchSize int, autoAck bool) (ds []amqp.Delivery) {
	ds = make([]amqp.Delivery, 0)

	// Get A Batch of Messages
GetBatchLoop:
	for {
		// Break if we have a full batch
		if len(ds) == batchSize {
			break GetBatchLoop
		}

		ch := qu.ex.rb.pool.GetChannelFromPool()

		d, ok, err := ch.Channel.Get(queueName, autoAck)
		// get data err, maybe connection/channel lost, retry
		if err != nil {
			qu.ex.rb.pool.ReturnChannel(ch, true)
			log.WithContext(qu.ex.rb.ops.ctx).WithError(err).Warn("failed to get batch, queue: %s, retry...", queueName)
			time.Sleep(time.Duration(qu.ex.rb.ops.healthCheckInterval) * time.Millisecond)
			continue
		}

		if !ok { // Break If empty
			break GetBatchLoop
		}

		ds = append(ds, d)
	}

	return
}
