package utils

import (
	"context"
	"testing"
)

func TestParseRedisURI(t *testing.T) {
	client, err := ParseRedisURI("redis://127.0.0.1:6379/0")
	if err != nil {
		panic(err)
	}
	t.Log(client.Ping(context.Background()))
}
