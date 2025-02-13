package rabbit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestQueue_Consume(t *testing.T) {
	ex := New(uri).
		Exchange(
			WithExchangeName("ex1"),
		)
	if ex.Error != nil {
		panic(ex.Error)
	}
	qu := ex.Queue(
		WithQueueName("q1"),
		WithQueueDeclare(false),
		WithQueueBind(false),
	)
	if qu.Error != nil {
		panic(qu.Error)
	}

	err := qu.Consume(
		handler,
		WithConsumeAutoRequestID(true),
	)
	if err != nil {
		t.Log(err)
	}
}

func handler(ctx context.Context, q string, delivery amqp.Delivery) bool {
	fmt.Println("handler", ctx, q, delivery.Exchange)
	return true
}

func TestQueue_ConsumeOne(t *testing.T) {
	rb := New(uri)
	if rb.Error != nil {
		panic(rb.Error)
	}
	for {
		time.Sleep(10 * time.Second)
		err := rb.
			Exchange(
				WithExchangeName("ex1"),
				WithExchangeDeclare(false),
			).Queue(
			WithQueueName("q1"),
			WithQueueDeclare(false),
			WithQueueBind(false),
		).ConsumeOne(
			10,
			handler,
		)
		t.Log(time.Now(), err)
	}
}
