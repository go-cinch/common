module github.com/go-cinch/common/id

go 1.20

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.0.3
	github.com/pkg/errors v0.9.1
	github.com/sony/sonyflake v1.1.0
)

require github.com/go-kratos/kratos/v2 v2.6.2 // indirect
