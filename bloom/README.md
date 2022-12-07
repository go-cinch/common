# Bloom Filter


simple bloom filter based on redis.


## Usage


```bash
go get -u github.com/go-cinch/common/bloom
```

```go
import (
	"fmt"
	"github.com/go-cinch/common/bloom"
	"github.com/go-redis/redis/v8"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	b := bloom.New(
		bloom.WithRedis(client),
	)

	b.Add("abc")
	fmt.Println(b.Exist("abc"))
	fmt.Println(b.Exist("def"))
	b.Add("def")
	fmt.Println(b.Exist("abc"))
	fmt.Println(b.Exist("def"))
}
```


## Options


- `WithRedis` - redis client, default 127.0.0.1:6379
- `WithKey` - redis cache key, default bloom
- `WithExpire` - key expire time, default 5 minutes
- `WithHash` - bloom hash function, default BKDRHash + SDBMHash + DJBHash
- `WithTimeout` - exec redis command timeout, default 5 seconds
