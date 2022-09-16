package utils

import (
	"context"
	"fmt"
	"testing"
)

func TestParseRedisURI(t *testing.T) {
	client, err := ParseRedisURI("redis://127.0.0.1:6379/0")
	if err != nil {
		panic(err)
	}
	fmt.Println(client.Ping(context.Background()))
}
