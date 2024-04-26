module github.com/go-cinch/common/rabbit

go 1.20

replace github.com/go-cinch/common/log => ../log

require (
	github.com/go-cinch/common/log v1.1.0
	github.com/google/uuid v1.3.1
	github.com/houseofcat/turbocookedrabbit/v2 v2.3.0
	github.com/pkg/errors v0.9.1
	github.com/streadway/amqp v1.1.0
	github.com/thoas/go-funk v0.9.3
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/Workiva/go-datastructures v1.1.0 // indirect
	github.com/go-kratos/kratos/v2 v2.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/crypto v0.10.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
