package jwt

import (
	"context"
	"github.com/go-cinch/common/copierx"
	jwtV4 "github.com/golang-jwt/jwt/v4"
	"github.com/golang-module/carbon/v2"
	"strings"
)

const (
	ClaimAuthorityId = "authorityId"
	ClaimExpires     = "exp"
)

type User struct {
	Code string
}

type user struct{}

func NewServerContext(ctx context.Context, claims jwtV4.Claims) context.Context {
	if mClaims, ok := claims.(jwtV4.MapClaims); ok {
		if v, ok2 := mClaims[ClaimAuthorityId].(string); ok2 {
			return NewServerContextByCode(ctx, v)
		}
	}
	return ctx
}

func NewServerContextByCode(ctx context.Context, code string) context.Context {
	c := strings.TrimSpace(code)
	if c != "" {
		ctx = context.WithValue(ctx, user{}, &User{Code: c})
	}
	return ctx
}

func FromServerContext(ctx context.Context) (u *User) {
	u = new(User)
	if v, ok := ctx.Value(user{}).(*User); ok {
		copierx.Copy(&u, v)
	}
	return
}

func (u *User) CreateToken(key, duration string) (token string, expires carbon.Carbon) {
	expires = carbon.Now().AddDuration(duration)
	claims := jwtV4.NewWithClaims(
		jwtV4.SigningMethodHS512,
		jwtV4.MapClaims{
			ClaimAuthorityId: u.Code,
			ClaimExpires:     expires.Timestamp(),
		},
	)
	token, _ = claims.SignedString([]byte(key))
	return
}
