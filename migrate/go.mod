module github.com/go-cinch/common/migrate/v2

go 1.25

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.2.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/lib/pq v1.10.9
	github.com/rubenv/sql-migrate v1.8.1
)

require (
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-kratos/kratos/v2 v2.8.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.38.0 // indirect
)
