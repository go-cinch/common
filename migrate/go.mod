module github.com/go-cinch/common/migrate

go 1.20

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.1.1
	github.com/go-sql-driver/mysql v1.7.1
	github.com/rubenv/sql-migrate v1.5.1
)

require (
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-kratos/kratos/v2 v2.7.0 // indirect
)
