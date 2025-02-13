package logging

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
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
			var stack string
			l := log.
				WithContext(ctx).
				WithFields(log.Fields{
					"kind": kind,
					"args": extractArgs(req),
				})
			defer func() {
				e := recover()
				if e != nil {
					buf := make([]byte, 64<<10) //nolint:gomnd
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					stack = fmt.Sprintf("%s", buf)
				}
				if stack != "" {
					l.WithFields(log.Fields{
						"stack": stack,
					}).Warn(operation)
				}
				if e != nil {
					err = errors.InternalServer("UNKNOWN", "unknown request error")
				}
			}()
			reply, err = handler(ctx, req)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			stack = extractError(err)
			l = l.WithFields(log.Fields{
				"resp":    extractReply(reply),
				"code":    code,
				"reason":  reason,
				"latency": time.Since(startTime).Seconds(),
			})
			if stack != "" {
				return
			}
			l.Info(operation)
			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
	if stringer, ok := req.(fmt.Stringer); ok {
		return truncate(stringer.String())
	}
	return truncate(fmt.Sprintf("%+v", req))
}

// extractReply returns the string of the reply
func extractReply(reply interface{}) string {
	if interfaceIsNil(reply) {
		return ""
	}
	if stringer, ok := reply.(fmt.Stringer); ok {
		return truncate(stringer.String())
	}
	return truncate(fmt.Sprintf("%+v", reply))
}

// extractError returns the string of the error
func extractError(err error) string {
	if err != nil {
		return truncate(fmt.Sprintf("%+v", err))
	}
	return ""
}

func truncate(s string) string {
	if len(s) <= 500 {
		return s
	}
	return s[:249] + "..." + s[len(s)-248:]
}

func interfaceIsNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return i == nil
}
