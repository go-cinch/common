module github.com/go-cinch/common/middleware/tenant/v2

go 1.25

replace (
	github.com/go-cinch/common/log => ../../log
	github.com/go-cinch/common/migrate/v2 => ../../migrate
	github.com/go-cinch/common/plugins/gorm/log => ../../plugins/gorm/log
	github.com/go-cinch/common/plugins/gorm/tenant/v2 => ../../plugins/gorm/tenant
)

require (
	github.com/go-cinch/common/plugins/gorm/tenant/v2 v2.0.1
	github.com/go-kratos/kratos/v2 v2.8.3
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-cinch/common/log v1.2.0 // indirect
	github.com/go-cinch/common/migrate/v2 v2.0.1 // indirect
	github.com/go-cinch/common/plugins/gorm/log v1.0.5 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-playground/form/v4 v4.2.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rubenv/sql-migrate v1.8.1 // indirect
	github.com/samber/lo v1.49.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
	gorm.io/gorm v1.31.1 // indirect
)
