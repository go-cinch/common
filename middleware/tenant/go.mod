module github.com/go-cinch/common/middleware/tenant

go 1.23

replace (
	github.com/go-cinch/common/log => ../../log
	github.com/go-cinch/common/migrate => ../../migrate
	github.com/go-cinch/common/plugins/gorm/log => ../../plugins/gorm/log
	github.com/go-cinch/common/plugins/gorm/tenant => ../../plugins/gorm/tenant
)

require (
	github.com/go-cinch/common/plugins/gorm/tenant v1.0.3
	github.com/go-kratos/kratos/v2 v2.8.3
)

require (
	github.com/go-cinch/common/log v1.2.0 // indirect
	github.com/go-cinch/common/migrate v1.0.5 // indirect
	github.com/go-cinch/common/plugins/gorm/log v1.0.5 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-playground/form/v4 v4.2.1 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rubenv/sql-migrate v1.5.1 // indirect
	github.com/samber/lo v1.49.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.1 // indirect
	gorm.io/gorm v1.25.2 // indirect
)
