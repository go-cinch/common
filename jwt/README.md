# Jwt

jwt token generator based on [golang-jwt](https://github.com/golang-jwt/jwt/v4), used
under [cinch layout](https://github.com/go-cinch/layout).

## Usage

```bash
go get -u github.com/go-cinch/common/jwt
```

- `NewServerContext` - new server context with jwt.Claims
- `NewServerContextByUser` - new server context with User
- `FromServerContext` - get user from server context
- `TokenFromServerContext` - get jwt token from server context
- `AppendToClientContext` - append user to grpc client
- `AppendToReplyHeader` - append user to grpc response header
- `User.CreateToken` - generate jwt token by User
