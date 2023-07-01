module github.com/go-cinch/common/plugins/gorm/log

go 1.20

replace github.com/go-cinch/common/log => ../../../log

require (
	github.com/go-cinch/common/log v1.0.3
	github.com/pkg/errors v0.9.1
	gorm.io/gorm v1.25.2
)

require (
	github.com/go-kratos/kratos/v2 v2.6.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
)
