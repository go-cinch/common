package migrate

import (
	"context"
	"embed"
	"reflect"
)

type Options struct {
	ctx         context.Context
	driver      string
	uri         string
	lockName    string
	before      func(ctx context.Context) error
	changeTable string
	fs          embed.FS
	fsRoot      string
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !interfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithDriver(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).driver = s
	}
}

func WithUri(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).uri = s
	}
}

func WithLockName(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lockName = s
	}
}

func WithBefore(f func(ctx context.Context) error) func(*Options) {
	return func(options *Options) {
		if f != nil {
			getOptionsOrSetDefault(options).before = f
		}
	}
}

func WithChangeTable(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).changeTable = s
	}
}

func WithFs(fs embed.FS) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).fs = fs
	}
}

func WithFsRoot(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).fsRoot = s
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			driver:      "mysql",
			uri:         "root:root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&timeout=10000ms",
			lockName:    "MigrationLock",
			changeTable: "schema_migrations",
		}
	}
	return options
}

func interfaceIsNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return i == nil
}
