package log

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	Info("test info")

	WithField("field1", 1).Info("test info with 1 field")
	WithField("field1", 1).Info("test info with 1 field and format %s", "yes")

	WithFields(Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field")
	WithFields(Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field and format %d", 1)

	WithError(fmt.Errorf("something error")).Info("test info with err")

	WithContext(context.WithValue(context.Background(), "ctx", "ctx")).Info("test info with ctx")
	WithContext(context.WithValue(context.Background(), "ctx", "ctx")).Info("test info with ctx and format %v", fmt.Errorf("error"))

	// override default wrapper
	DefaultWrapper = NewWrapper(
		WithLogger(
			log.With(
				log.NewStdLogger(os.Stdout),
				// logrus example: https://github.com/go-kratos/kratos/tree/main/contrib/log/logrus
				// kratos "github.com/go-kratos/kratos/contrib/log/logrus/v2"
				// "github.com/sirupsen/logrus"
				// kratos.NewLogger(logrus.New()),
				"ts", log.DefaultTimestamp,
				"caller", log.DefaultCaller,
				"service.id", "123",
				"service.name", "456",
				"service.version", "789",
				"trace.id", tracing.TraceID(),
				"span.id", tracing.SpanID(),
			),
		),
		WithLoggerMessageKey("m"),
	)

	Info("test info")

	WithField("field1", 1).Info("test info with 1 field")
	WithField("field1", 1).Info("test info with 1 field and format %s", "yes")

	WithFields(Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field")
	WithFields(Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field and format %d", 1)

	WithContext(context.WithValue(context.Background(), "ctx", "ctx")).Info("test info with ctx")
	WithContext(context.WithValue(context.Background(), "ctx", "ctx")).Info("test info with ctx and format %v", fmt.Errorf("error"))

	DefaultWrapper = NewWrapper(
		WithLevel(ErrorLevel),
	)
	Debug("test debug")
	Info("test info")
	Warn("test warn")
	Error("test error")
	Fatal("test fatal")
}
