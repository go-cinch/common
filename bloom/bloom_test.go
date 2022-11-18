package bloom

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	b := New(WithRedis(client))
	i := 10000000000
	arr := make([]string, 0)
	for i < 10001000000 {
		arr = append(arr, fmt.Sprintf("%d", i))
		if len(arr) == 100000 {
			err := b.Add(arr...)
			fmt.Println(i, time.Now(), err)
			arr = make([]string, 0)
		}
		i++
	}
	j := 10000999000
	for j < 10001000999 {
		if b.Exist(fmt.Sprintf("%d", j)) {
			fmt.Printf("%d possible exist\n", j)
		} else {
			fmt.Printf("%d not exist\n", j)
		}
		j++
	}
	fmt.Println("end")
}
