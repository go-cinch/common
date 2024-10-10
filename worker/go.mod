module github.com/go-cinch/common/worker

go 1.22

toolchain go1.22.3

replace (
	github.com/go-cinch/common/log => ../log
	github.com/go-cinch/common/nx => ../nx
)

require (
	github.com/go-cinch/common/log v1.1.1
	github.com/go-cinch/common/nx v1.0.4
	github.com/golang-module/carbon/v2 v2.3.12
	github.com/google/uuid v1.3.1
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	github.com/hibiken/asynq v0.24.1
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.2.1
	go.opentelemetry.io/otel v1.30.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-kratos/kratos/v2 v2.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	go.opentelemetry.io/otel/metric v1.30.0 // indirect
	go.opentelemetry.io/otel/trace v1.30.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/grpc v1.56.3 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)
