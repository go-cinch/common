module github.com/go-cinch/common/captcha

go 1.20

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.0.3
	github.com/mojocn/base64Captcha v1.3.5
	github.com/redis/go-redis/v9 v9.0.5
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-kratos/kratos/v2 v2.6.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	golang.org/x/image v0.0.0-20190501045829-6d32002ffd75 // indirect
)
