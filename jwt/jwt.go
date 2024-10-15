package jwt

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	jwtV4 "github.com/golang-jwt/jwt/v4"
	"github.com/golang-module/carbon/v2"
)

const (
	ClaimExpire = "exp"
)

type User struct {
	Token string            `json:"token"`
	Attrs map[string]string `json:"attrs"`
}

type user struct{}

func NewServerContext(ctx context.Context, claims jwtV4.Claims, keys ...string) context.Context {
	mClaims, ok := claims.(jwtV4.MapClaims)
	if !ok {
		return ctx
	}
	u := &User{
		Attrs: make(map[string]string),
	}
	for _, key := range keys {
		v, ok2 := mClaims[key].(string)
		if !ok2 {
			continue
		}
		u.Attrs[key] = v
	}
	ctx = NewServerContextByUser(ctx, *u)
	return ctx
}

func NewServerContextByUser(ctx context.Context, u User) context.Context {
	ctx = context.WithValue(ctx, user{}, &u)
	return ctx
}

func FromServerContext(ctx context.Context) (u *User) {
	u = &User{
		Attrs: make(map[string]string),
	}
	if v, ok := ctx.Value(user{}).(*User); ok {
		u.Token = v.Token
		// copy attr
		for k2, v2 := range v.Attrs {
			u.Attrs[k2] = v2
		}
	} else if md, ok2 := metadata.FromServerContext(ctx); ok2 {
		for k2, v2 := range md {
			if strings.HasPrefix(k2, "x-md-global-") {
				u.Attrs[strings.TrimPrefix(k2, "x-md-global-")] = v2[0]
			}
		}
	}
	if u.Token == "" {
		u.Token = TokenFromServerContext(ctx)
	}
	return
}

func TokenFromServerContext(ctx context.Context) (token string) {
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

func AppendToReplyHeader(ctx context.Context, us ...User) {
	var u *User
	if len(us) > 0 {
		u = &us[0]
	} else {
		u = FromServerContext(ctx)
	}
	tr, _ := transport.FromServerContext(ctx)
	for k, v := range u.Attrs {
		tr.ReplyHeader().Set("x-md-global-"+k, v)
	}
	return
}

func (u *User) CreateToken(key, duration string) (token string, expire carbon.Carbon) {
	expire = carbon.Now().AddDuration(duration)
	mClaims := jwtV4.MapClaims{
		ClaimExpire: expire.Timestamp(),
	}
	for k, v := range u.Attrs {
		mClaims[k] = v
	}
	claims := jwtV4.NewWithClaims(
		jwtV4.SigningMethodHS512,
		mClaims,
	)
	token, _ = claims.SignedString([]byte(key))
	return
}
