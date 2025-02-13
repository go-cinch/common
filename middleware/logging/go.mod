module github.com/go-cinch/common/middleware/logging

go 1.23

toolchain go1.23.4

replace github.com/go-cinch/common/log => ../../log

require (
	github.com/go-cinch/common/log v1.2.0
	github.com/go-kratos/kratos/v2 v2.8.3
)

require (
	github.com/go-playground/form/v4 v4.2.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/grpc v1.61.1 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
