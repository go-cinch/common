module github.com/go-cinch/common/idempotent

go 1.23

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.2.0
	github.com/google/uuid v1.4.0
	github.com/redis/go-redis/v9 v9.7.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-kratos/kratos/v2 v2.8.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
)
