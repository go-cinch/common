package mq

import (
	"testing"
)

const uri = "amqp://guest:guest@127.0.0.1:5672/"

func TestNew(t *testing.T) {
	rb := New(uri)
	if rb.Error != nil {
		panic(rb.Error)
	}

	ex := rb.Exchange(
		WithExchangeName("ex1"),
	)
	if ex.Error != nil {
		panic(ex.Error)
	}
	err := ex.QueueWithDeadLetter(
		WithQueueName("q1"),
		WithQueueRouteKeys("rt1"),
		WithQueueDeadLetterName("dl-ex"),
		WithQueueDeadLetterKey("dlr"),
		WithQueueMessageTTL(30000),
	).Error
	if err != nil {
		panic(err)
	}

	err = ex.QueueWithDeadLetter(
		WithQueueName("q2"),
		WithQueueRouteKeys("rt2"),
		WithQueueDeadLetterName("dl-ex"),
		WithQueueDeadLetterKey("dlr"),
		WithQueueMessageTTL(30000),
	).Error
	if err != nil {
		panic(err)
	}

	err = rb.Exchange(
		WithExchangeName("ex2"),
	).Queue(
		WithQueueName("q3"),
		WithQueueRouteKeys("rt3"),
	).Error
	if err != nil {
		panic(err)
	}

	err = rb.Exchange(
		WithExchangeName("dl-ex"),
	).Queue(
		WithQueueName("dlq"),
		WithQueueRouteKeys("dlr"),
	).Error
}
