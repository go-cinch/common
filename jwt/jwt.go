package jwt

import (
	"context"
	"github.com/go-cinch/common/copierx"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	jwtV4 "github.com/golang-jwt/jwt/v4"
	"github.com/golang-module/carbon/v2"
	gmd "google.golang.org/grpc/metadata"
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

func NewServerContextByReplyMD(ctx context.Context, md gmd.MD) context.Context {
	u := new(User)
	v1 := md.Get("x-md-global-code")
	if len(v1) == 1 {
		u.Code = v1[0]
	}
	v2 := md.Get("x-md-global-platform")
	if len(v2) == 1 {
		u.Platform = v2[0]
	}
	return NewServerContextByUser(ctx, *u)
}

func FromServerContext(ctx context.Context) (u *User) {
	u = new(User)
	if v, ok := ctx.Value(user{}).(*User); ok {
		copierx.Copy(&u, v)
	} else if md, ok2 := metadata.FromServerContext(ctx); ok2 {
		u.Code = md.Get("x-md-global-code")
		u.Platform = md.Get("x-md-global-platform")
		u.Token = TokenFromServerContext(ctx)
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

func AppendToClientContext(ctx context.Context, us ...User) context.Context {
	var u *User
	if len(us) > 0 {
		u = &us[0]
	} else {
		u = FromServerContext(ctx)
	}
	ctx = metadata.AppendToClientContext(
		ctx,
		"x-md-global-code", u.Code,
		"x-md-global-platform", u.Platform,
		"x-md-global-jwt", u.Token,
	)
	return ctx
}

func AppendToReplayHeader(ctx context.Context, us ...User) {
	var u *User
	if len(us) > 0 {
		u = &us[0]
	} else {
		u = FromServerContext(ctx)
	}
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr.ReplyHeader() != nil {
			tr.ReplyHeader().Set("x-md-global-code", u.Code)
			tr.ReplyHeader().Set("x-md-global-platform", u.Platform)
		}
	}
	return
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
