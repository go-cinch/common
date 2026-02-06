package migrate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-cinch/common/log"
	migrate "github.com/go-cinch/common/migrate/v2"
	"github.com/go-sql-driver/mysql"
	mysqlDriver "gorm.io/driver/mysql"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Migrate struct {
	ops Options
	db  *gorm.DB
}

func New(options ...func(*Options)) (*Migrate, error) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.dsn == "" {
		return nil, fmt.Errorf("dsn is required")
	}
	m := &Migrate{
		ops: *ops,
	}
	showDsn := getShowDsn(ops.driver, ops.dsn)
	if ops.skipMigrate {
		log.
			WithField("dsn", showDsn).
			Info("open db...")
		db, err := openDBWithPool(ops, ops.dsn)
		if err != nil {
			log.
				WithField("dsn", showDsn).
				Info("open db failed")
			return nil, err
		}
		m.db = db
		log.
			WithField("dsn", showDsn).
			Info("open db success")
	}
	return m, nil
}

func (m *Migrate) Migrate() error {
	if m.ops.skipMigrate {
		return nil
	}
	showDsn := getShowDsn(m.ops.driver, m.ops.dsn)
	log.
		WithField("dsn", showDsn).
		Info("migrating...")
	db, err := m.migrate(m.ops.dsn)
	if err != nil {
		log.
			WithField("dsn", showDsn).
			Info("migrate failed")
		return err
	}
	m.db = db
	log.
		WithField("dsn", showDsn).
		Info("migrate success")
	return nil
}

func (m *Migrate) DB() *gorm.DB {
	return m.db.Session(&gorm.Session{})
}

func (m *Migrate) DBWithContext(ctx context.Context) *gorm.DB {
	return m.db.Session(&gorm.Session{}).WithContext(ctx)
}

func (m *Migrate) migrate(dsn string) (db *gorm.DB, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = migrate.Do(
		migrate.WithCtx(ctx),
		migrate.WithDriver(m.ops.driver),
		migrate.WithURI(dsn),
		migrate.WithFs(m.ops.sqlFile),
		migrate.WithFsRoot(m.ops.sqlRoot),
		migrate.WithBefore(func(_ context.Context) (err error) {
			db, err = openDBWithPool(&m.ops, dsn)
			return
		}),
	)
	return
}

func openDB(driver, dsn string, config *gorm.Config) (*gorm.DB, error) {
	switch driver {
	case "mysql":
		return gorm.Open(mysqlDriver.Open(dsn), config)
	case "postgres":
		return gorm.Open(postgresDriver.Open(dsn), config)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

func openDBWithPool(ops *Options, dsn string) (*gorm.DB, error) {
	db, err := openDB(ops.driver, dsn, ops.config)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(ops.maxIdle)
	sqlDB.SetMaxOpenConns(ops.maxOpen)
	sqlDB.SetConnMaxLifetime(ops.maxLifetime)
	return db, nil
}

func getShowDsn(driver, dsn string) string {
	switch driver {
	case "mysql":
		cfg, e := mysql.ParseDSN(dsn)
		if e == nil {
			cfg.Passwd = "***"
			return cfg.FormatDSN()
		}
	case "postgres":
		return hidePostgresPassword(dsn)
	}
	return ""
}

func hidePostgresPassword(dsn string) string {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		if idx := strings.Index(dsn, "://"); idx != -1 {
			prefix := dsn[:idx+3]
			rest := dsn[idx+3:]
			if atIdx := strings.Index(rest, "@"); atIdx != -1 {
				userPass := rest[:atIdx]
				hostAndRest := rest[atIdx:]
				if colonIdx := strings.Index(userPass, ":"); colonIdx != -1 {
					user := userPass[:colonIdx]
					return prefix + user + ":***" + hostAndRest
				}
			}
		}
		return dsn
	}
	pairs := strings.Fields(dsn)
	result := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 && kv[0] == "password" {
			result = append(result, "password=***")
		} else {
			result = append(result, pair)
		}
	}
	return strings.Join(result, " ")
}

