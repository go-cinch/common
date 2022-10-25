package jwt

import (
	"context"
	"github.com/go-cinch/common/copierx"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	jwtV4 "github.com/golang-jwt/jwt/v4"
	"github.com/golang-module/carbon/v2"
	"strings"
)

const (
	ClaimCode     = "code"
	ClaimPlatform = "platform"
	ClaimExpires  = "exp"
)

type User struct {
	Token    string
	Code     string
	Platform string
}

type user struct{}

func NewServerContext(ctx context.Context, claims jwtV4.Claims) context.Context {
	if mClaims, ok := claims.(jwtV4.MapClaims); ok {
		u := new(User)
		if v, ok2 := mClaims[ClaimCode].(string); ok2 {
			u.Code = v
		}
		if v, ok3 := mClaims[ClaimPlatform].(string); ok3 {
			u.Platform = v
		}
		ctx = NewServerContextByUser(ctx, *u)
	}
	return ctx
}

func NewServerContextByUser(ctx context.Context, u User) context.Context {
	ctx = context.WithValue(ctx, user{}, &u)
	return ctx
}

func FromServerContext(ctx context.Context) (u *User) {
	u = new(User)
	if v, ok := ctx.Value(user{}).(*User); ok {
		copierx.Copy(&u, v)
	}
	return
}

func FromClientContext(ctx context.Context) (u *User) {
	u = new(User)
	if md, ok := metadata.FromServerContext(ctx); ok {
		u.Code = md.Get("x-md-global-code")
		u.Platform = md.Get("x-md-global-platform")
	}
	u.Token = TokenFromClientContext(ctx)
	return
}

func TokenFromClientContext(ctx context.Context) (token string) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		auths := strings.SplitN(tr.RequestHeader().Get("Authorization"), " ", 2)
		if len(auths) == 2 && strings.EqualFold(auths[0], "Bearer") {
			token = auths[1]
			return
		}
	}
	if md, ok := metadata.FromServerContext(ctx); ok {
		token = md.Get("x-md-global-jwt")
		return
	}
	return
}

func AppendToClientContext(ctx context.Context) context.Context {
	u := FromClientContext(ctx)
	ctx = metadata.AppendToClientContext(
		ctx,
		"x-md-global-code", u.Code,
		"x-md-global-platform", u.Platform,
		"x-md-global-jwt", u.Token,
	)
	return ctx
}

func (u *User) CreateToken(key, duration string) (token string, expires carbon.Carbon) {
	expires = carbon.Now().AddDuration(duration)
	claims := jwtV4.NewWithClaims(
		jwtV4.SigningMethodHS512,
		jwtV4.MapClaims{
			ClaimCode:     u.Code,
			ClaimPlatform: u.Platform,
			ClaimExpires:  expires.Timestamp(),
		},
	)
	token, _ = claims.SignedString([]byte(key))
	return
}
