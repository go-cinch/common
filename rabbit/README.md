# Rabbit


rabbitmq connection pool based on [amqp](https://github.com/streadway/amqp) and [turbocookedrabbit](https://github.com/houseofcat/turbocookedrabbit).


## Usage


```bash
go get -u github.com/go-cinch/common/rabbit
```

```go
import (
	"context"
	"fmt"
	"github.com/go-cinch/common/rabbit"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	// 1. init rabbit instance
	rb := rabbit.New(
		"amqp://guest:guest@127.0.0.1:5672/",
		rabbit.WithName("my app"),
	)

	err := rb.Error
	if err != nil {
		panic(err)
	}

	// 2.1. get exchange and declare if not exist
	ex := rb.Exchange(
		rabbit.WithExchangeName("exchange1"),
	)
	err = ex.Error
	if err != nil {
		panic(err)
	}

	// 2.2. get exchange without declare
	ex = rb.Exchange(
		rabbit.WithExchangeName("exchange1"),
		rabbit.WithExchangeDeclare(false),
	)

	// 3.1. send json msg to exchange
	err = ex.PublishJson("{}")
	if err != nil {
		panic(err)
	}

	// 3.2. send proto msg to exchange
	var p emptypb.Empty
	err = ex.PublishProto(&p)
	if err != nil {
		panic(err)
	}

	// 3.3. send byte msg to exchange
	err = ex.PublishByte([]byte("ok"))
	if err != nil {
		panic(err)
	}

	// 4.1. get queue and declare if not exist
	q1 := ex.Queue(
		rabbit.WithQueueName("q1"),
	)
	err = q1.Error
	if err != nil {
		panic(err)
	}

	// 4.2. get queue without declare
	q1 = ex.Queue(
		rabbit.WithQueueName("q1"),
		rabbit.WithQueueDeclare(false),
	)

	// 4.3. get queue without declare, and bind route keys
	q1 = ex.Queue(
		rabbit.WithQueueName("q1"),
		rabbit.WithQueueDeclare(false),
		rabbit.WithQueueBind(true),
		rabbit.WithQueueRouteKeys("key1", "key2"),
	)

	// 4.4. declare queue q2 with dead letter queue dlx
	ex.QueueWithDeadLetter(
		rabbit.WithQueueName("q2"),
		rabbit.WithQueueRouteKeys("key1"),
		rabbit.WithQueueDeadLetterName("dlx"),
		rabbit.WithQueueDeadLetterKey("key1.dead"),
		rabbit.WithQueueMessageTTL(30000),
	)

	// 4.4. consume queue 50 pieces of data
	err = q1.ConsumeOne(50, consumer)
	if err != nil {
		panic(err)
	}

	// 4.5 consume handle
	err = q1.Consume(consumer)
	if err != nil {
		panic(err)
	}
}

func consumer(ctx context.Context, q string, d amqp.Delivery) bool {
	fmt.Println(ctx, q)
	return true
}
```


## Options


### Rabbit
  

- `WithCtx` - context
- `WithName` - app name  
- `WithHeartbeat` - pool heartbeat, default 1 second
- `WithTimeout` - connection timeout, default 10 second
- `WithMaxChannel` - max channel count, default 50
- `WithMaxConnection` - max connection count, default 10
- `WithHealthCheckInterval` - healthcheck interval, default 100 milli second


### ExchangeOptions

- `WithExchangeName` - name
- `WithExchangeKind` - kind, default direct 
- `WithExchangeDurable` - durable, default true 
- `WithExchangeDeclare` - declare, default true
- `WithExchangeAutoDelete` - auto delete, default false 
- `WithExchangeInternal` - internal exchange, do not accept publishings, default false   
- `WithExchangeNoWait` - declare without waiting for a confirmation from the server, default false  
- `WithExchangeNamePrefix` - exchange name prefix 
- `WithExchangeArgs` - other args 


### QueueOptions  

- `WithQueueName` - name
- `WithQueueDurable` - durable, default true 
- `WithQueueDeclare` - declare, default true  
- `WithQueueBind` - bind route key, default true 
- `WithQueueRouteKeys` - route key array  
- `WithQueueAutoDelete` - auto delete, default false 
- `WithQueueExclusive` - the server will ensure that this is the sole consumer from this queue, default false 
- `WithQueueNoWait` - declare without waiting for a confirmation from the server, default false 
- `WithQueueArgs` - other args  
- `WithQueueBindArgs` - other bind args 
- `WithQueueDeadLetterName` - dead letter name
- `WithQueueDeadLetterKey` - dead letter route key 
- `WithQueueMessageTTL` - msg expiration 
   

### PublishOptions  


- `WithPublishCtx` - context
- `WithPublishMaxRetryCount` - max retry count, default 3
- `WithPublishTimeout` - timeout, default 10000 milli second
- `WithPublishReconnectInterval` - reconnect interval, default 1000 milli second
- `WithPublishIdleInterval` - idle interval, default 1000 milli second
- `WithPublishRouteKey` - route key
- `WithPublishContentType` - content type
- `WithPublishHeaders` - headers
- `WithPublishDeadLetter` - publish to dead letter queue, default false
- `WithPublishDeadLetterFirstQueue` - original queue name

### ConsumeOptions

- `WithConsumeQosPrefetchCount` - qos, default 2
- `WithConsumeConsumer` - consumer name 
- `WithConsumeAutoAck` - auto confirm, default false 
- `WithConsumeExclusive` - exclusive, default false 
- `WithConsumeNoWait` - do not wait for the server to confirm the request and immediately begin deliveries, default false 
- `WithConsumeNackRequeue` - requeue when delivery ack is false, default false
- `WithConsumeNackRetry` - retry when delivery ack is false, default false 
- `WithConsumeNackMaxRetryCount` - max retry count when NackRetry is true, default 5
- `WithConsumeAutoRequestId` - auth generate request id 
- `WithConsumeOneContext` - consume one ctx  
- `WithConsumeArgs` - other args 
