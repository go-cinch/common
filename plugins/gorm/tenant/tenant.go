package tenant

import (
	"context"
	"errors"
	"time"

	"github.com/go-cinch/common/log"
	"github.com/go-cinch/common/migrate"
	"github.com/go-cinch/common/utils"
	kratosLog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-sql-driver/mysql"
	m "gorm.io/driver/mysql"
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
	tenantIds []string
	tenantDBs map[string]*gorm.DB
}

func New(options ...func(*Options)) (*Tenant, error) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	tenantIds := make([]string, 0, len(ops.dsn))
	for k := range ops.dsn {
		if !utils.Contains(tenantIds, k) {
			tenantIds = append(tenantIds, k)
		}
	}
	if len(tenantIds) == 0 {
		return nil, errors.New("at least one tenant")
	}
	tenantDBs := make(map[string]*gorm.DB)
	if ops.skipMigrate {
		for _, id := range tenantIds {
			dsn := ops.dsn[id]
			showDsn := getShowDsn(dsn)
			log.
				WithField("id", id).
				Info("tenant open db...")
			db, err := gorm.Open(m.Open(dsn), ops.config)
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
		tenantIds: tenantIds,
		tenantDBs: tenantDBs,
	}, nil
}

func (t *Tenant) Migrate() error {
	for _, id := range t.tenantIds {
		dsn := t.ops.dsn[id]
		showDsn := getShowDsn(dsn)
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
		v = t.tenantDBs[t.tenantIds[0]]
	}
	return v.Session(&gorm.Session{}).WithContext(ctx)
}

func (t *Tenant) migrate(dsn string) (db *gorm.DB, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = migrate.Do(
		migrate.WithCtx(ctx),
		migrate.WithUri(dsn),
		migrate.WithFs(t.ops.sqlFile),
		migrate.WithFsRoot(t.ops.sqlRoot),
		migrate.WithBefore(func(ctx context.Context) (err error) {
			db, err = gorm.Open(m.Open(dsn), t.ops.config)
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
			// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
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

func getShowDsn(dsn string) string {
	cfg, e := mysql.ParseDSN(dsn)
	if e == nil {
		// hidden password
		cfg.Passwd = "***"
		return cfg.FormatDSN()
	}
	return ""
}
