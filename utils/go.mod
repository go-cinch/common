module github.com/go-cinch/common/utils

go 1.20

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/r3labs/diff/v3 v3.0.1
	github.com/redis/go-redis/v9 v9.2.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-kratos/kratos/v2 v2.7.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
