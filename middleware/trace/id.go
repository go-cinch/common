package trace

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel/trace"
)

func ID() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
				if tr, ok := transport.FromServerContext(ctx); ok {
					tr.ReplyHeader().Set("x-trace-id", span.TraceID().String())
				}
			}
			return handler(ctx, req)
		}
	}
}
