package migrate

import (
	"context"
	"embed"
	"fmt"
	"testing"
)

//go:embed db/*.sql
var sqlFs embed.FS

func TestDoMySQL(t *testing.T) {
	err := Do(
		WithCtx(context.WithValue(context.Background(), "k", "v")),
		WithDriver("mysql"),
		WithURI("root:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&timeout=10000ms"),
		WithFs(sqlFs),
		WithFsRoot("db"),
		WithBefore(beforeMigrate),
	)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("MySQL migration completed successfully")
}

func TestDoPostgres(t *testing.T) {
	// Test with key=value format
	err := Do(
		WithCtx(context.WithValue(context.Background(), "k", "v")),
		WithDriver("postgres"),
		WithURI("host=localhost user=root password=password dbname=test port=5432 sslmode=disable"),
		WithFs(sqlFs),
		WithFsRoot("db"),
		WithBefore(beforeMigrate),
	)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("PostgreSQL migration completed successfully")
}

func TestDoPostgresURL(t *testing.T) {
	// Test with URL format
	err := Do(
		WithCtx(context.WithValue(context.Background(), "k", "v")),
		WithDriver("postgres"),
		WithURI("postgres://root:password@localhost:5432/test?sslmode=disable"),
		WithFs(sqlFs),
		WithFsRoot("db"),
		WithBefore(beforeMigrate),
	)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("PostgreSQL migration (URL format) completed successfully")
}

func beforeMigrate(ctx context.Context) (err error) {
	fmt.Println(ctx)
	return
}
