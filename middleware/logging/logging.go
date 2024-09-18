package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/go-cinch/common/log"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Server is an server logging middleware.
func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}
			reply, err = handler(ctx, req)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			stack := extractError(err)
			l := log.
				WithContext(ctx).
				WithFields(log.Fields{
					"kind":    kind,
					"args":    extractArgs(req),
					"code":    code,
					"latency": time.Since(startTime).Seconds(),
				})
			if stack != "" {
				l.
					WithFields(log.Fields{
						"reason": reason,
						"stack":  stack,
					}).
					Warn(operation)
			} else {
				l.Info(operation)
			}
			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}

// extractError returns the string of the error
func extractError(err error) string {
	if err != nil {
		return fmt.Sprintf("%+v", err)
	}
	return ""
}
