package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-cinch/common/log"
	migrate "github.com/go-cinch/common/migrate/v2"
	kratosLog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-sql-driver/mysql"
	"github.com/samber/lo"
	mysqlDriver "gorm.io/driver/mysql"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ID() kratosLog.Valuer {
	return func(ctx context.Context) interface{} {
		return FromContext(ctx)
	}
}

type tenantCtx struct{}

func NewContext(ctx context.Context, id string) context.Context {
	ctx = context.WithValue(ctx, tenantCtx{}, id)
	return ctx
}

func FromContext(ctx context.Context) (id string) {
	if v, ok := ctx.Value(tenantCtx{}).(string); ok {
		id = v
	}
	return
}

type Tenant struct {
	ops       Options
	tenantIDs []string
	tenantDBs map[string]*gorm.DB
}

func New(options ...func(*Options)) (*Tenant, error) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	tenantIDs := make([]string, 0, len(ops.dsn))
	for k := range ops.dsn {
		if !lo.Contains(tenantIDs, k) {
			tenantIDs = append(tenantIDs, k)
		}
	}
	if len(tenantIDs) == 0 {
		return nil, errors.New("at least one tenant")
	}
	tenantDBs := make(map[string]*gorm.DB)
	if ops.skipMigrate {
		for _, id := range tenantIDs {
			dsn := ops.dsn[id]
			showDsn := getShowDsn(ops.driver, dsn)
			log.
				WithField("id", id).
				Info("tenant open db...")
			db, err := openDB(ops.driver, dsn, ops.config)
			if err != nil {
				log.
					WithField("id", id).
					WithField("dsn", showDsn).
					Info("tenant open db failed")
				return nil, err
			}
			tenantDBs[id] = db
			log.
				WithField("id", id).
				WithField("dsn", showDsn).
				Info("tenant open db success")
		}
	}
	return &Tenant{
		ops:       *ops,
		tenantIDs: tenantIDs,
		tenantDBs: tenantDBs,
	}, nil
}

func (t *Tenant) Migrate() error {
	for _, id := range t.tenantIDs {
		dsn := t.ops.dsn[id]
		showDsn := getShowDsn(t.ops.driver, dsn)
		log.
			WithField("id", id).
			Info("tenant migrating...")
		db, err := t.migrate(t.ops.dsn[id])
		if err != nil {
			log.
				WithField("id", id).
				WithField("dsn", showDsn).
				Info("tenant migrate failed")
			return err
		}
		t.tenantDBs[id] = db
		log.
			WithField("id", id).
			WithField("dsn", showDsn).
			Info("tenant migrate success")
	}
	return nil
}

func (t *Tenant) DB(ctx context.Context) *gorm.DB {
	id := FromContext(ctx)
	v, ok := t.tenantDBs[id]
	if !ok {
		// invalid tenant id use default 0
		v = t.tenantDBs[t.tenantIDs[0]]
	}
	return v.Session(&gorm.Session{}).WithContext(ctx)
}

func (t *Tenant) migrate(dsn string) (db *gorm.DB, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = migrate.Do(
		migrate.WithCtx(ctx),
		migrate.WithDriver(t.ops.driver),
		migrate.WithURI(dsn),
		migrate.WithFs(t.ops.sqlFile),
		migrate.WithFsRoot(t.ops.sqlRoot),
		migrate.WithBefore(func(_ context.Context) (err error) {
			db, err = openDB(t.ops.driver, dsn, t.ops.config)
			if err != nil {
				return
			}
			// fix: packets.go:122: closing bad idle connection: EOF
			// https://gorm.io/docs/generic_interface.html#Connection-Pool
			// Get generic database object sql.DB to use its functions
			sqlDB, err := db.DB()
			if err != nil {
				return
			}
			// SetMaxIDleConns sets the maximum number of connections in the idle connection pool.
			sqlDB.SetMaxIdleConns(t.ops.maxIdle)
			// SetMaxOpenConns sets the maximum number of open connections to the database.
			sqlDB.SetMaxOpenConns(t.ops.maxOpen)
			// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
			sqlDB.SetConnMaxLifetime(t.ops.maxLifetime)
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

func getShowDsn(driver, dsn string) string {
	switch driver {
	case "mysql":
		cfg, e := mysql.ParseDSN(dsn)
		if e == nil {
			// hidden password
			cfg.Passwd = "***"
			return cfg.FormatDSN()
		}
	case "postgres":
		return hidePostgresPassword(dsn)
	}
	return ""
}

func hidePostgresPassword(dsn string) string {
	// Handle URL format: postgres://user:password@host:port/dbname?params
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// Find password section
		if idx := strings.Index(dsn, "://"); idx != -1 {
			prefix := dsn[:idx+3]
			rest := dsn[idx+3:]

			// Find @ symbol
			if atIdx := strings.Index(rest, "@"); atIdx != -1 {
				userPass := rest[:atIdx]
				hostAndRest := rest[atIdx:]

				// Replace password
				if colonIdx := strings.Index(userPass, ":"); colonIdx != -1 {
					user := userPass[:colonIdx]
					return prefix + user + ":***" + hostAndRest
				}
			}
		}
		return dsn
	}

	// Handle key=value format: host=localhost user=postgres password=secret dbname=test
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
