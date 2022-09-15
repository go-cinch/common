package migrate

import (
	"context"
	"embed"
	"fmt"
	"testing"
)

//go:embed db/*.sql
var sqlFs embed.FS

func TestDo(t *testing.T) {
	err := Do(
		WithCtx(context.WithValue(context.Background(), "k", "v")),
		WithUri("root:root@tcp(127.0.0.1:4306)/test?charset=utf8mb4&parseTime=True&timeout=10000ms"),
		WithFs(sqlFs),
		WithFsRoot("db"),
		WithBefore(beforeMigrate),
	)
	fmt.Println(err)
}

func beforeMigrate(ctx context.Context) (err error) {
	fmt.Println(ctx)
	return
}
