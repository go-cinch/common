# Log


simple log wrapper based on [kratos log](https://go-kratos.dev/en/docs/component/log).


## Characteristic


- `Caller` - auto get line number
- `Format` - print normal str or format str by Info/Warn/Error, no need Infof/Warnf/Errorf
- `WithContext` - support [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) log tracking by WithContext
- `WithError` - print err field by WithError
- `WithField` - print custom field by WithField/WithFields
- `Custom Plugin` - support add custom log plugin, such as logrus/zap


## Usage


```bash
go get -u github.com/go-cinch/common/log
```


```go
import (
	"context"
	"fmt"
	"github.com/go-cinch/common/log"
	kratos "github.com/go-kratos/kratos/contrib/log/logrus/v2"
	kratosLog "github.com/go-kratos/kratos/v2/log"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	// 1. use default wrapper
	log.Info("test info")

	log.WithField("field1", 1).Info("test info with 1 field")
	log.WithField("field1", 1).Info("test info with 1 field and format %s", "yes")

	log.WithFields(log.Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field")
	log.WithFields(log.Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field and format %d", 1)

	log.WithContext(context.WithValue(context.Background(), "ctx", "ctx")).Info("test info with ctx")
	log.WithContext(context.WithValue(context.Background(), "ctx", "ctx")).Info("test info with ctx and format %v", fmt.Errorf("error"))

	// 2. override default wrapper
	log.DefaultWrapper = log.NewWrapper(
		log.WithLogger(
			kratosLog.With(
				kratosLog.NewStdLogger(os.Stdout),
				"service.id", "123",
				"service.name", "456",
				"service.version", "789",
			),
		),
		log.WithLevel(log.WarnLevel),
		log.WithLoggerMessageKey("m"),
	)
	// only print test warn
	log.Info("test info")
	log.Warn("test warn")

	// 3. use logrus
	// go get -u github.com/go-kratos/kratos/contrib/log/logrus/v2
	// go get -u github.com/sirupsen/logrus
	log.DefaultWrapper = log.NewWrapper(
		log.WithLogger(
			kratosLog.With(
				kratos.NewLogger(logrus.New()),
			),
		),
	)
	log.Info("test info")

	
	// INFO caller=main.go:15 msg=test info
	// INFO caller=main.go:17 msg=test info with 1 field field1=1
	// INFO caller=main.go:18 msg=test info with 1 field and format yes field1=1
	// INFO caller=main.go:20 msg=test info with 2 field field1=1 filed2=2
	// INFO caller=main.go:21 msg=test info with 2 field and format 1 field1=1 filed2=2
	// INFO caller=main.go:23 msg=test info with ctx
	// INFO caller=main.go:24 msg=test info with ctx and format error
	// WARN service.id=123 service.name=456 service.version=789 caller=main.go:41 m=test warn
	// INFO[0000] test info                                     caller="main.go:57"
}
```


## Options


- `WithLevel - log level, default debug
- `WithLogger` - kratos logger, default kratosLog.DefaultLogger
- `WithLoggerMessageKey` - msg key, default msg
