package i18n

import (
	"context"
	"github.com/go-cinch/common/i18n"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/text/language"
	"google.golang.org/grpc/metadata"
)

var i = i18n.New()

type translator struct{}

func Translator(options ...func(*i18n.Options)) middleware.Middleware {
	i = i18n.New(options...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (rp interface{}, err error) {
			var lang language.Tag
			header := make(metadata.MD)
			key := "accept-language"
			if tr, ok := transport.FromServerContext(ctx); ok {
				accept := tr.RequestHeader().Get(key)
				lang = language.Make(accept)
			}
			ii := i.Select(lang)
			header.Set(key, ii.Language().String())
			ctx = metadata.NewOutgoingContext(ctx, header)
			ctx = NewContext(ctx, ii)
			return handler(ctx, req)
		}
	}
}

func NewContext(ctx context.Context, i *i18n.I18n) context.Context {
	ctx = context.WithValue(ctx, translator{}, i)
	return ctx
}

func FromContext(ctx context.Context) (rp *i18n.I18n) {
	rp = i
	if v, ok := ctx.Value(translator{}).(*i18n.I18n); ok {
		rp = v
	}
	return
}
