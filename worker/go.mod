module github.com/go-cinch/common/worker

go 1.23

replace (
	github.com/go-cinch/common/log => ../log
	github.com/go-cinch/common/queue/stream => ../queue/stream
)

require (
	github.com/bsm/redislock v0.9.4
	github.com/go-cinch/common/log v1.2.0
	github.com/go-cinch/common/queue/stream v1.0.0
	github.com/golang-module/carbon/v2 v2.3.12
	github.com/google/uuid v1.6.0
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	github.com/hibiken/asynq v0.25.1
	github.com/paulbellamy/ratecounter v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.7.0
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/trace v1.34.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-kratos/kratos/v2 v2.8.3 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)
