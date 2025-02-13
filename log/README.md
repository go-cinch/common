# Log

simple log wrapper based on [logrus](https://github.com/sirupsen/logrus).

## Characteristic

- `Caller` - auto get line number
- `Format` - print normal str or format str by Info/Warn/Error, no need Infof/Warnf/Errorf
- `WithContext` - support [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) log tracking by
  WithContext
- `WithError` - print err field by WithError
- `WithField` - print custom field by WithField/WithFields
- `Custom Plugin` - support add custom log plugin, such as logrus/zap

## Usage

```bash
go get github.com/go-cinch/common/log
```

```go
import (
	"context"
	"fmt"
	
	"github.com/go-cinch/common/log"
	"github.com/go-cinch/common/log/caller"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
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
	logOps := []func(*log.Options){
		log.WithJSON(false),
		log.WithLevel(log.WarnLevel),
		log.WithValuer("custom.field", "ok"),
		log.WithValuer("trace.id", tracing.TraceID()),
		log.WithValuer("span.id", tracing.SpanID()),
		log.WithSkipEmpty(false),
		log.WithCallerOptions(
			caller.WithSource(false),
			caller.WithLevel(2),
			caller.WithVersion(true),
		),
	}
	log.DefaultWrapper = log.NewWrapper(logOps...)
	// only print test warn
	log.Info("test info(not print)")
	log.Warn("test warn")
}
```

## Options

- `WithLevel - log level, default debug
