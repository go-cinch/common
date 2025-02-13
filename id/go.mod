module github.com/go-cinch/common/id

go 1.23

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sony/sonyflake v1.1.0
)

require (
	github.com/go-kratos/kratos/v2 v2.8.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
)
