package tenant

import (
	"embed"
	"time"

	"github.com/go-cinch/common/plugins/gorm/log"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Options struct {
	driver      string
	dsn         map[string]string
	sqlFile     embed.FS
	sqlRoot     string
	config      *gorm.Config
	maxIdle     int
	maxOpen     int
	maxLifetime time.Duration
	skipMigrate bool
}

func WithDriver(driver string) func(*Options) {
	return func(options *Options) {
		if driver == "mysql" || driver == "postgres" {
			getOptionsOrSetDefault(options).driver = driver
		}
	}
}

func WithDSN(tenant, dsn string) func(*Options) {
	return func(options *Options) {
		data := getOptionsOrSetDefault(options)
		// Validate DSN based on driver
		if getShowDsn(data.driver, dsn) != "" {
			if _, ok := data.dsn[tenant]; !ok {
				data.dsn[tenant] = dsn
			}
		}
	}
}

func WithSQLFile(fs embed.FS) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).sqlFile = fs
	}
}

func WithSQLRoot(root string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).sqlRoot = root
	}
}

func WithConfig(config *gorm.Config) func(*Options) {
	return func(options *Options) {
		if config != nil {
			getOptionsOrSetDefault(options).config = config
		}
	}
}

func WithMaxIdle(count int) func(*Options) {
	return func(options *Options) {
		if count > 0 {
			getOptionsOrSetDefault(options).maxIdle = count
		}
	}
}

func WithMaxOpen(count int) func(*Options) {
	return func(options *Options) {
		if count > 0 {
			getOptionsOrSetDefault(options).maxOpen = count
		}
	}
}

func WithMaxLifetime(d time.Duration) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).maxLifetime = d
	}
}

func WithSkipMigrate(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).skipMigrate = flag
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			driver:  "postgres",
			dsn:     make(map[string]string),
			sqlRoot: "migrations",
			config: &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					SingularTable: true,
				},
				QueryFields: true,
				Logger: log.New(
					log.WithColorful(true),
					log.WithSlow(200),
				),
			},
			maxIdle:     10,
			maxOpen:     100,
			maxLifetime: time.Hour,
		}
	}
	return options
}
