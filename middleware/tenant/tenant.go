package tenant

import (
	"context"
	"github.com/go-cinch/common/plugins/gorm/tenant"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

func Tenant() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (rp interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				tenantId := tr.RequestHeader().Get("x-tenant-id")
				ctx = tenant.NewContext(ctx, tenantId)
			}
			return handler(ctx, req)
		}
	}
}
