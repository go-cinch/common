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
		WithURI("root:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&timeout=10000ms"),
		WithFs(sqlFs),
		WithFsRoot("db"),
		WithBefore(beforeMigrate),
	)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log()
}

func beforeMigrate(ctx context.Context) (err error) {
	fmt.Println(ctx)
	return
}
