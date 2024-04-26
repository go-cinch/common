package log

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-cinch/common/log/caller"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

func getCtx() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil {
			return ""
		}
		if v, ok := ctx.Value("ctx").(string); ok {
			return v
		}
		return ""
	}
}

func TestDefault(*testing.T) {
	Info("test info")

	WithField("field1", 1).Info("test info with 1 field")
	WithField("field1", 1).Info("test info with 1 field and format %s", "yes")

	WithFields(Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field")
	WithFields(Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field and format %d", 1)

	WithError(fmt.Errorf("something error")).Warn("test warn with err")
	WithError(nil).Warn("test warn without err")
	WithError(fmt.Errorf("something error")).Error("test error with err")
	WithError(nil).Error("test error without err")

	WithContext(context.WithValue(context.Background(), "ctx", "ctx")).Info("test info with ctx")
	WithContext(context.WithValue(context.Background(), "ctx", "ctx")).Info("test info with ctx and format %v", fmt.Errorf("error"))

	// override default wrapper
	DefaultWrapper = NewWrapper(
		WithLevel(InfoLevel),
		WithJSON(true),
		WithValuer("service.id", "id"),
		WithValuer("service.name", "name"),
		WithValuer("service.version", "1.0.0"),
		WithValuer("trace.id", tracing.TraceID()),
		WithValuer("span.id", tracing.SpanID()),
		WithValuer("ctx", getCtx()),
		WithCallerOptions(
			caller.WithSource(false),
			caller.WithLevel(2),
			caller.WithVersion(true),
		),
	)

	Info("test info")

	WithField("field1", 1).Info("test info with 1 field")
	WithField("field1", 1).Info("test info with 1 field and format %s", "yes")

	WithFields(Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field")
	WithFields(Fields{"field1": 1, "filed2": 2}).Info("test info with 2 field and format %d", 1)

	WithContext(context.WithValue(context.Background(), "ctx", "ctx1")).Info("test info with ctx")
	WithContext(context.WithValue(context.Background(), "ctx", "ctx2")).Info("test info with ctx and format %v", fmt.Errorf("error"))

	DefaultWrapper = NewWrapper(
		WithLevel(ErrorLevel),
	)
	Debug("test debug, not print since error level")
	Info("test info, not print since error level")
	Warn("test warn, not print since error level")
	Error("test error")
	Fatal("test fatal")
	Info("not print since os.Exit()")
}
