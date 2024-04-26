module github.com/go-cinch/common/plugins/gorm/log

go 1.20

replace github.com/go-cinch/common/log => ../../../log

require (
	github.com/go-cinch/common/log v1.1.0
	github.com/pkg/errors v0.9.1
	gorm.io/gorm v1.25.2
)

require (
	github.com/go-kratos/kratos/v2 v2.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
