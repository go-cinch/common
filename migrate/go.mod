module github.com/go-cinch/common/migrate

go 1.23

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.2.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/rubenv/sql-migrate v1.5.1
)

require (
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-kratos/kratos/v2 v2.8.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
)
