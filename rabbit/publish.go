package mq

import (
	"github.com/google/uuid"
	"github.com/houseofcat/turbocookedrabbit/v2/pkg/tcr"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
	"sync/atomic"
	"time"
)

type Publish struct {
	ops       PublishOptions
	ex        *Exchange
	msg       amqp.Publishing
	publisher *tcr.Publisher
	Error     error
}

// PublishProto publish grpc proto msg
func (ex *Exchange) PublishProto(m proto.Message, options ...func(*PublishOptions)) (err error) {
	var b []byte
	b, err = proto.Marshal(m)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = ex.PublishByte(b, options...)
	return
}

// PublishJson publish str msg
func (ex *Exchange) PublishJson(m string, options ...func(*PublishOptions)) (err error) {
	err = ex.PublishByte([]byte(m), options...)
	return
}

// PublishByte publish byte msg
func (ex *Exchange) PublishByte(m []byte, options ...func(*PublishOptions)) (err error) {
	if len(m) == 0 {
		err = errors.Errorf("msg is empty")
		return
	}
	pu := ex.beforePublish(options...)
	if pu.Error != nil {
		err = errors.WithStack(pu.Error)
		return
	}
	pu.msg.Body = m
	err = pu.publish()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (ex *Exchange) beforePublish(options ...func(*PublishOptions)) *Publish {
	var pu Publish
	if ex.Error != nil {
		pu.Error = ex.Error
		return &pu
	}
	ops := getPublishOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if len(ops.routeKeys) == 0 {
		pu.Error = errors.Errorf("route key is empty")
		return &pu
	}
	if ops.deadLetter {
		if ops.deadLetterFirstQueue == "" {
			pu.Error = errors.Errorf("dead letter first queue is empty")
			return &pu
		}
		if _, ok := ops.headers["x-retry-count"].(int32); !ok {
			ops.headers["x-retry-count"] = 0
		}
		ops.headers["x-first-death-queue"] = ops.deadLetterFirstQueue
	}
	if ops.deliveryMode <= 0 || ops.deliveryMode > amqp.Persistent {
		ops.deliveryMode = amqp.Persistent
	}
	pu.ops = *ops
	msg := amqp.Publishing{
		DeliveryMode: ops.deliveryMode,
		Timestamp:    time.Now(),
		ContentType:  ops.contentType,
		Headers:      ops.headers,
		Expiration:   ops.expiration,
	}
	pu.msg = msg
	pu.ex = ex
	pu.publisher = tcr.NewPublisherFromConfig(
		&tcr.RabbitSeasoning{
			PoolConfig: ex.rb.poolConfig,
			PublisherConfig: &tcr.PublisherConfig{
				AutoAck:                false,
				SleepOnIdleInterval:    uint32(ops.idleInterval),
				SleepOnErrorInterval:   uint32(ops.reconnectInterval),
				PublishTimeOutInterval: uint32(ops.timeout),
			},
		},
		pu.ex.rb.pool,
	)
	return &pu
}

func (pu *Publish) publish() (err error) {
	if atomic.LoadInt32(&pu.ex.rb.lost) == 1 {
		err = errors.Errorf("connection maybe lost")
		return
	}
	for _, key := range pu.ops.routeKeys {
		envelope := &tcr.Envelope{
			DeliveryMode: pu.msg.DeliveryMode,
			Exchange:     pu.ex.ops.name,
			RoutingKey:   key,
			ContentType:  pu.msg.ContentType,
			Headers:      pu.msg.Headers,
			Mandatory:    pu.ops.mandatory,
			Immediate:    pu.ops.immediate,
		}
		letter := &tcr.Letter{
			LetterID:   uuid.New(),
			RetryCount: uint32(pu.ops.maxRetryCount),
			Body:       pu.msg.Body,
			Envelope:   envelope,
		}
		err = pu.publisher.PublishWithConfirmationContextError(
			pu.ops.ctx,
			letter,
		)
		if err != nil {
			err = errors.Wrapf(err, "publish failed")
			return
		}
	}
	return
}
