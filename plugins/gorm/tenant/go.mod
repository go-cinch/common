module github.com/go-cinch/common/plugins/gorm/tenant

go 1.23

replace (
	github.com/go-cinch/common/log => ../../../log
	github.com/go-cinch/common/migrate => ../../../migrate
	github.com/go-cinch/common/plugins/gorm/log => ../../../plugins/gorm/log
)

require (
	github.com/go-cinch/common/log v1.2.0
	github.com/go-cinch/common/migrate v1.0.5
	github.com/go-cinch/common/plugins/gorm/log v1.0.5
	github.com/go-kratos/kratos/v2 v2.8.3
	github.com/go-sql-driver/mysql v1.7.1
	github.com/samber/lo v1.49.1
	gorm.io/driver/mysql v1.5.1
	gorm.io/gorm v1.25.2
)

require (
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rubenv/sql-migrate v1.5.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)
