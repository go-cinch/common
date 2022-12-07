package rabbit

import (
	"context"
	"github.com/streadway/amqp"
	"github.com/thoas/go-funk"
	"reflect"
)

type Options struct {
	ctx                 context.Context
	name                string
	heartbeat           int
	timeout             int
	maxConnection       int
	maxChannel          int
	healthCheckInterval int
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !interfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithName(name string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).name = name
	}
}

func WithHeartbeat(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).heartbeat = second
		}
	}
}

func WithTimeout(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).timeout = second
		}
	}
}

func WithMaxConnection(count int) func(*Options) {
	return func(options *Options) {
		if count > 0 {
			getOptionsOrSetDefault(options).maxConnection = count
		}
	}
}

func WithMaxChannel(count int) func(*Options) {
	return func(options *Options) {
		if count > 0 {
			getOptionsOrSetDefault(options).maxChannel = count
		}
	}
}

func WithHealthCheckInterval(milli int) func(*Options) {
	return func(options *Options) {
		if milli > 0 {
			getOptionsOrSetDefault(options).healthCheckInterval = milli
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			ctx:                 context.Background(),
			heartbeat:           1,
			timeout:             10,
			maxConnection:       10,
			maxChannel:          50,
			healthCheckInterval: 100,
		}
	}
	return options
}

type ExchangeOptions struct {
	name       string
	kind       string
	durable    bool
	autoDelete bool
	internal   bool
	noWait     bool
	args       amqp.Table
	declare    bool
	namePrefix string
}

func WithExchangeName(name string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).name = name
	}
}

func WithExchangeKind(kind string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).kind = kind
	}
}

func WithExchangeDurable(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).durable = flag
	}
}

func WithExchangeAutoDelete(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).autoDelete = flag
	}
}

func WithExchangeInternal(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).internal = flag
	}
}

func WithExchangeNoWait(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).noWait = flag
	}
}

func WithExchangeArgs(args amqp.Table) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).args = args
	}
}

func WithExchangeDeclare(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).declare = flag
	}
}

func WithExchangeNamePrefix(prefix string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).namePrefix = prefix
	}
}

func getExchangeOptionsOrSetDefault(options *ExchangeOptions) *ExchangeOptions {
	if options == nil {
		return &ExchangeOptions{
			kind:    amqp.ExchangeDirect,
			durable: true,
			declare: true,
		}
	}
	return options
}

type QueueOptions struct {
	name           string
	routeKeys      []string
	durable        bool
	autoDelete     bool
	exclusive      bool
	noWait         bool
	args           amqp.Table
	bindArgs       amqp.Table
	declare        bool
	bind           bool
	deadLetterName string
	deadLetterKey  string
	messageTTL     int32
	namePrefix     string
}

func WithQueueName(name string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).name = name
	}
}

func WithQueueRouteKeys(keys ...string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).routeKeys = append(getQueueOptionsOrSetDefault(options).routeKeys, keys...)
	}
}

func WithQueueDurable(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).durable = flag
	}
}

func WithQueueAutoDelete(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).autoDelete = flag
	}
}

func WithQueueExclusive(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).exclusive = flag
	}
}

func WithQueueNoWait(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).noWait = flag
	}
}

func WithQueueArgs(args amqp.Table) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).args = args
	}
}

func WithQueueBindArgs(args amqp.Table) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).bindArgs = args
	}
}

func WithQueueDeclare(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).declare = flag
	}
}

func WithQueueBind(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).bind = flag
	}
}

func WithQueueDeadLetterName(name string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).deadLetterName = name
	}
}

func WithQueueDeadLetterKey(key string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).deadLetterKey = key
	}
}

func WithQueueMessageTTL(ttl int32) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).messageTTL = ttl
	}
}

func getQueueOptionsOrSetDefault(options *QueueOptions) *QueueOptions {
	if options == nil {
		return &QueueOptions{
			durable: true,
			declare: true,
			bind:    true,
		}
	}
	return options
}

type PublishOptions struct {
	ctx                  context.Context
	maxRetryCount        int
	timeout              int
	reconnectInterval    int
	idleInterval         int
	routeKeys            []string
	contentType          string
	headers              amqp.Table
	deliveryMode         uint8
	mandatory            bool
	immediate            bool
	expiration           string
	deadLetter           bool
	deadLetterFirstQueue string
}

func WithPublishCtx(ctx context.Context) func(*PublishOptions) {
	return func(options *PublishOptions) {
		if !interfaceIsNil(ctx) {
			getPublishOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithPublishMaxRetryCount(count int) func(*PublishOptions) {
	return func(options *PublishOptions) {
		if count > 0 {
			getPublishOptionsOrSetDefault(options).maxRetryCount = count
		}
	}
}

func WithPublishTimeout(milli int) func(*PublishOptions) {
	return func(options *PublishOptions) {
		if milli > 0 {
			getPublishOptionsOrSetDefault(options).timeout = milli
		}
	}
}

func WithPublishReconnectInterval(milli int) func(*PublishOptions) {
	return func(options *PublishOptions) {
		if milli > 0 {
			getPublishOptionsOrSetDefault(options).reconnectInterval = milli
		}
	}
}

func WithPublishIdleInterval(milli int) func(*PublishOptions) {
	return func(options *PublishOptions) {
		if milli > 0 {
			getPublishOptionsOrSetDefault(options).idleInterval = milli
		}
	}
}

func WithPublishRouteKey(keys ...string) func(*PublishOptions) {
	return func(options *PublishOptions) {
		d := getPublishOptionsOrSetDefault(options)
		for _, item := range keys {
			if !funk.ContainsString(d.routeKeys, item) {
				d.routeKeys = append(d.routeKeys, item)
			}
		}
	}
}

func WithPublishContentType(contentType string) func(*PublishOptions) {
	return func(options *PublishOptions) {
		getPublishOptionsOrSetDefault(options).contentType = contentType
	}
}

func WithPublishHeaders(headers amqp.Table) func(*PublishOptions) {
	return func(options *PublishOptions) {
		if headers != nil {
			getPublishOptionsOrSetDefault(options).headers = headers
		}
	}
}

func WithPublishDeadLetter(flag bool) func(*PublishOptions) {
	return func(options *PublishOptions) {
		getPublishOptionsOrSetDefault(options).deadLetter = flag
	}
}

func WithPublishDeadLetterFirstQueue(q string) func(*PublishOptions) {
	return func(options *PublishOptions) {
		getPublishOptionsOrSetDefault(options).deadLetterFirstQueue = q
	}
}

func getPublishOptionsOrSetDefault(options *PublishOptions) *PublishOptions {
	if options == nil {
		return &PublishOptions{
			ctx:               context.Background(),
			contentType:       "text/plain",
			headers:           amqp.Table{},
			maxRetryCount:     3,
			timeout:           10000,
			reconnectInterval: 1000,
			idleInterval:      1000,
		}
	}
	return options
}

type ConsumeOptions struct {
	qosPrefetchCount  int
	consumer          string
	autoAck           bool
	exclusive         bool
	noWait            bool
	args              amqp.Table
	nackRequeue       bool
	nackRetry         bool
	nackMaxRetryCount int32
	autoRequestId     bool
	oneCtx            context.Context
}

func WithConsumeQosPrefetchCount(prefetchCount int) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).qosPrefetchCount = prefetchCount
	}
}

func WithConsumeConsumer(consumer string) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).consumer = consumer
	}
}

func WithConsumeAutoAck(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).autoAck = flag
	}
}

func WithConsumeExclusive(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).exclusive = flag
	}
}

func WithConsumeNoWait(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).noWait = flag
	}
}

func WithConsumeArgs(args amqp.Table) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).args = args
	}
}

func WithConsumeNackRequeue(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).nackRequeue = flag
	}
}

func WithConsumeNackRetry(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).nackRetry = flag
	}
}

func WithConsumeNackMaxRetryCount(i int32) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		if i > 0 {
			getConsumeOptionsOrSetDefault(options).nackMaxRetryCount = i
		}
	}
}

func WithConsumeAutoRequestId(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).autoRequestId = flag
	}
}

func WithConsumeOneContext(ctx context.Context) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).oneCtx = ctx
	}
}

func getConsumeOptionsOrSetDefault(options *ConsumeOptions) *ConsumeOptions {
	if options == nil {
		return &ConsumeOptions{
			qosPrefetchCount:  2,
			nackMaxRetryCount: 5,
		}
	}
	return options
}

func interfaceIsNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return i == nil
}
