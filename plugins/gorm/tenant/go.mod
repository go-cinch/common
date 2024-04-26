module github.com/go-cinch/common/plugins/gorm/tenant

go 1.20

replace (
	github.com/go-cinch/common/log => ../../../log
	github.com/go-cinch/common/migrate => ../../../migrate
	github.com/go-cinch/common/plugins/gorm/log => ../../../plugins/gorm/log
	github.com/go-cinch/common/utils => ../../../utils
)

require (
	github.com/go-cinch/common/log v1.1.0
	github.com/go-cinch/common/migrate v1.0.4
	github.com/go-cinch/common/plugins/gorm/log v1.0.4
	github.com/go-cinch/common/utils v1.0.4
	github.com/go-kratos/kratos/v2 v2.7.0
	github.com/go-sql-driver/mysql v1.7.1
	gorm.io/driver/mysql v1.5.1
	gorm.io/gorm v1.25.2
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/r3labs/diff/v3 v3.0.1 // indirect
	github.com/redis/go-redis/v9 v9.2.1 // indirect
	github.com/rubenv/sql-migrate v1.5.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
