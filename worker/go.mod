module github.com/go-cinch/common/worker

go 1.20

replace (
	github.com/go-cinch/common/log => ../log
	github.com/go-cinch/common/nx => ../nx
)

require (
	github.com/go-cinch/common/log v1.0.3
	github.com/go-cinch/common/nx v1.0.3
	github.com/golang-module/carbon/v2 v2.2.3
	github.com/google/uuid v1.3.0
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	github.com/hibiken/asynq v0.24.1
	github.com/pkg/errors v0.9.1
	github.com/redis/go-redis/v9 v9.0.5
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-kratos/kratos/v2 v2.6.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
