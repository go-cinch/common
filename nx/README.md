# Nx


simple nx lock based on redis.


## Usage


```bash
go get -u github.com/go-cinch/common/nx
```

```
import (
	"context"
	"fmt"
	"github.com/go-cinch/common/nx"
	"github.com/go-redis/redis/v8"
	"time"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	n := nx.New(
		nx.WithRedis(client),
		nx.WithKey("nx.lock.example"),
	)

	// Lock && Unlock example
	fmt.Println("first lock", n.Lock(context.Background()))
	i := 0
	for i < 5 {
		// retry lock
		fmt.Println("retry lock", n.Lock())
		time.Sleep(500 * time.Millisecond)
		i++
	}
	n.Unlock()
	fmt.Println("after unlock", n.Lock(context.Background()))
	n.Unlock()
	// first lock true
	// retry lock false
	// retry lock false
	// retry lock false
	// retry lock false
	// retry lock false
	// after unlock true

	// MustLock && Unlock example
	fmt.Println("get lock", time.Now(), n.Lock(context.Background()))
	go func() {
		time.Sleep(3 * time.Second)
		n.Unlock()
		fmt.Println("unlock", time.Now())
	}()

	go func() {
		err := n.MustLock(context.Background())
		if err != nil {
			fmt.Println("must lock failed", time.Now())
			return
		}
		fmt.Println("must lock success", time.Now())
	}()

	time.Sleep(10 * time.Second)
}
```


## Options


- `WithRedis` - redis client, default 127.0.0.1:6379
- `WithKey` - redis cache key, default nx.lock
- `WithExpire` - key expire time, default 1 minute, avoid deadlock, it should not be set too long


## Caution


avoid deadlock, `MustLock` will auto retry 400 times to get lock in 10s, if failed, u will get an error
