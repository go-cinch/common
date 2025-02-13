package i18n

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-cinch/common/i18n"
	"github.com/go-kratos/kratos/v2/errors"
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

func NewError(ctx context.Context, text string, f func(string, ...interface{}) *errors.Error, args ...string) error {
	text = FromContext(ctx).T(text)
	if len(args) == 0 {
		return f(text)
	}
	formats := make([]string, 0, len(args))
	vs := make([]interface{}, 0, len(args))
	for index := 0; index < len(args); index += 2 {
		if index+1 < len(args) {
			formats = append(formats, "%s: %s")
			vs = append(vs, args[index], args[index+1])
			continue
		}
		formats = append(formats, "%s")
		vs = append(vs, args[index])
	}
	return f(fmt.Sprintf("%s %s", text, fmt.Sprintf(strings.Join(formats, ", "), vs...)))
}
