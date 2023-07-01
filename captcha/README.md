# Captcha


base64 captcha otp based on redis and [base64Captcha](https://github.com/mojocn/base64Captcha).


## Usage


```bash
go get -u github.com/go-cinch/common/captcha
```

```go
import (
	"context"
	"fmt"
	"github.com/go-cinch/common/captcha"
	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	c := captcha.New(
		captcha.WithRedis(client),
		captcha.WithCtx(context.Background()),
	)
	// get captcha id and base64 img
	id, img := c.Get()
	fmt.Println(id, img)

	// verify captcha by str and id
	fmt.Println(c.Verify(id, "1234"))
}
```


## Options


- `WithRedis` - redis client, default 127.0.0.1:6379
- `WithCtx` - context, convenient log tracking
- `WithPrefix` - redis cache key prefix, default captcha_
- `WithExpire` - key expire time, default 5 minutes
- `WithNum` - number of characters, default 4
