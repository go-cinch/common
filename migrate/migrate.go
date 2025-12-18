package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-cinch/common/log"
	m "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

func Do(options ...func(*Options)) (err error) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}

	err = database(ops)
	if err != nil {
		return
	}

	var db *sql.DB
	db, err = sql.Open(ops.driver, ops.uri)
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("open %s(%s) failed", ops.driver, ops.uri)
		return
	}

	defer func() {
		releaseErr := releaseLock(ops, db)
		if releaseErr != nil && err == nil {
			err = releaseErr
		}
	}()

	var lockAcquired bool
	for {
		lockAcquired, err = acquireLock(ops, db)
		if err != nil {
			return
		}
		if lockAcquired {
			break
		}
		log.
			WithContext(ops.ctx).
			WithFields(log.Fields{
				"migrate.lock": ops.lockName,
			}).
			Info("cannot acquire advisory lock, retrying...")
	}

	if ops.before != nil {
		err = ops.before(ops.ctx)
		if err != nil {
			log.
				WithContext(ops.ctx).
				WithError(err).
				Error("exec before callback failed")
			return
		}
	}

	rollback := os.Getenv("SQL_MIGRATE_ROLLBACK")
	if rollback != "" {
		log.
			WithContext(ops.ctx).
			WithField("sql", rollback).
			Info("exec rollback")
		arr := strings.Split(rollback, "; ")
		for i, item := range arr {
			if strings.TrimSpace(item) == "" {
				continue
			}
			_, err = db.ExecContext(ops.ctx, item)
			if err != nil {
				log.
					WithContext(ops.ctx).
					WithError(err).
					WithField(strings.Join([]string{"sql", strconv.Itoa(i + 1)}, "."), item).
					Error("exec rollback failed")
				return
			}
		}
		log.
			WithContext(ops.ctx).
			WithField("sql", rollback).
			Info("exec rollback success")
	}

	migrate.SetTable(ops.changeTable)
	source := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: ops.fs,
		Root:       ops.fsRoot,
	}
	err = status(ops, db, source)
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("show migrate status failed")
		return
	}

	_, err = migrate.Exec(db, ops.driver, source, migrate.Up)
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("migrate failed")
		return
	}
	log.
		WithContext(ops.ctx).
		Info("migrate success")
	return
}

func database(ops *Options) (err error) {
	switch ops.driver {
	case "mysql":
		return createMySQLDatabase(ops)
	case "postgres":
		return createPostgresDatabase(ops)
	default:
		err = fmt.Errorf("unsupported driver: %s", ops.driver)
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("unsupported database driver")
		return
	}
}

func createMySQLDatabase(ops *Options) (err error) {
	var cfg *m.Config
	cfg, err = m.ParseDSN(ops.uri)
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("invalid mysql uri")
		return
	}
	dbname := cfg.DBName
	cfg.DBName = ""
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return
	}
	defer db.Close()
	_, err = db.Exec(strings.Join([]string{"CREATE DATABASE IF NOT EXISTS `", dbname, "`"}, ""))
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("create mysql database failed")
	}
	return
}

func createPostgresDatabase(ops *Options) (err error) {
	// Parse PostgreSQL DSN to extract database name
	dbname := extractPostgresDBName(ops.uri)
	if dbname == "" {
		log.
			WithContext(ops.ctx).
			Warn("no database name found in postgres uri, skip database creation")
		return
	}

	// Connect to postgres default database to create target database
	dsnWithoutDB := replacePostgresDBName(ops.uri, "postgres")
	db, err := sql.Open("postgres", dsnWithoutDB)
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("failed to connect to postgres")
		return
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname).Scan(&exists)
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("failed to check postgres database existence")
		return
	}

	if !exists {
		// Create database (cannot use parameterized query for CREATE DATABASE)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			log.
				WithContext(ops.ctx).
				WithError(err).
				Error("create postgres database failed")
			return
		}
		log.
			WithContext(ops.ctx).
			WithField("database", dbname).
			Info("postgres database created")
	}
	return
}

func extractPostgresDBName(dsn string) string {
	// Try URL format first: postgres://user:password@localhost:5432/dbname?params
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parts := strings.Split(dsn, "/")
		if len(parts) >= 4 {
			dbPart := parts[3]
			// Remove query parameters
			if idx := strings.Index(dbPart, "?"); idx != -1 {
				return dbPart[:idx]
			}
			return dbPart
		}
	}

	// Try key=value format: host=localhost user=postgres dbname=test port=5432
	pairs := strings.Fields(dsn)
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 && kv[0] == "dbname" {
			return kv[1]
		}
	}
	return ""
}

func replacePostgresDBName(dsn, newDBName string) string {
	// Handle URL format
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parts := strings.Split(dsn, "/")
		if len(parts) >= 4 {
			// Replace database name
			dbPart := parts[3]
			if idx := strings.Index(dbPart, "?"); idx != -1 {
				parts[3] = newDBName + dbPart[idx:]
			} else {
				parts[3] = newDBName
			}
			return strings.Join(parts, "/")
		}
	}

	// Handle key=value format
	pairs := strings.Fields(dsn)
	result := make([]string, 0, len(pairs))
	found := false
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 && kv[0] == "dbname" {
			result = append(result, fmt.Sprintf("dbname=%s", newDBName))
			found = true
		} else {
			result = append(result, pair)
		}
	}
	if !found {
		result = append(result, fmt.Sprintf("dbname=%s", newDBName))
	}
	return strings.Join(result, " ")
}

func hashLockName(name string) int64 {
	// Simple hash function to convert string to int64 for PostgreSQL advisory lock
	var hash int64
	for i := 0; i < len(name); i++ {
		hash = hash*31 + int64(name[i])
	}
	return hash
}

func acquireLock(ops *Options, db *sql.DB) (f bool, err error) {
	switch ops.driver {
	case "mysql":
		return acquireMySQLLock(ops, db)
	case "postgres":
		return acquirePostgresLock(ops, db)
	default:
		err = fmt.Errorf("unsupported driver: %s", ops.driver)
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("unsupported database driver for lock")
		return
	}
}

func acquireMySQLLock(ops *Options, db *sql.DB) (f bool, err error) {
	// GET_LOCK will be blocked if another session already acquired the lock
	// timeout 5s
	q := strings.Join([]string{"SELECT GET_LOCK('", "', 5)"}, ops.lockName)
	err = db.QueryRow(q).Scan(&f)

	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			WithFields(log.Fields{
				"migrate.lock": ops.lockName,
			}).
			Error("acquire mysql advisory lock for migration failed")
		return
	}

	log.
		WithContext(ops.ctx).
		WithFields(log.Fields{
			"migrate.lock": ops.lockName,
		}).
		Info("acquire mysql advisory lock: %v", f)
	return
}

func acquirePostgresLock(ops *Options, db *sql.DB) (f bool, err error) {
	// Use pg_try_advisory_lock which returns immediately
	// Convert lock name to a numeric key using hash
	lockKey := hashLockName(ops.lockName)
	err = db.QueryRow("SELECT pg_try_advisory_lock($1)", lockKey).Scan(&f)

	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			WithFields(log.Fields{
				"migrate.lock": ops.lockName,
				"lock.key":     lockKey,
			}).
			Error("acquire postgres advisory lock for migration failed")
		return
	}

	log.
		WithContext(ops.ctx).
		WithFields(log.Fields{
			"migrate.lock": ops.lockName,
			"lock.key":     lockKey,
		}).
		Info("acquire postgres advisory lock: %v", f)
	return
}

func releaseLock(ops *Options, db *sql.DB) (err error) {
	switch ops.driver {
	case "mysql":
		return releaseMySQLLock(ops, db)
	case "postgres":
		return releasePostgresLock(ops, db)
	default:
		err = fmt.Errorf("unsupported driver: %s", ops.driver)
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("unsupported database driver for lock release")
		return
	}
}

func releaseMySQLLock(ops *Options, db *sql.DB) (err error) {
	q := strings.Join([]string{"SELECT RELEASE_LOCK('", "')"}, ops.lockName)
	_, err = db.Exec(q)

	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			WithFields(log.Fields{
				"migrate.lock": ops.lockName,
			}).
			Error("release mysql advisory lock for migration failed")
		return err
	}

	log.
		WithContext(ops.ctx).
		WithFields(log.Fields{
			"migrate.lock": ops.lockName,
		}).
		Info("release mysql advisory lock success")
	return
}

func releasePostgresLock(ops *Options, db *sql.DB) (err error) {
	lockKey := hashLockName(ops.lockName)
	_, err = db.Exec("SELECT pg_advisory_unlock($1)", lockKey)

	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			WithFields(log.Fields{
				"migrate.lock": ops.lockName,
				"lock.key":     lockKey,
			}).
			Error("release postgres advisory lock for migration failed")
		return err
	}

	log.
		WithContext(ops.ctx).
		WithFields(log.Fields{
			"migrate.lock": ops.lockName,
			"lock.key":     lockKey,
		}).
		Info("release postgres advisory lock success")
	return
}

func status(ops *Options, db *sql.DB, source *migrate.EmbedFileSystemMigrationSource) (err error) {
	var migrations []*migrate.Migration
	migrations, err = source.FindMigrations()
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("find migration failed")
		return
	}

	var records []*migrate.MigrationRecord
	records, err = migrate.GetMigrationRecords(db, ops.driver)
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("find migration history failed")
		return
	}
	rows := make(map[string]bool)
	pending := make([]string, 0)
	applied := make([]string, 0)
	for _, item := range migrations {
		rows[item.Id] = false
	}

	for _, item := range records {
		rows[item.Id] = true
	}

	for i, l := 0, len(migrations); i < l; i++ {
		if !rows[migrations[i].Id] {
			pending = append(pending, migrations[i].Id)
		} else {
			applied = append(applied, migrations[i].Id)
		}
	}
	log.
		WithContext(ops.ctx).
		WithFields(log.Fields{
			"migrate.pending": strings.Join(pending, ","),
			"migrate.applied": strings.Join(applied, ","),
		}).
		Info("migration status, pending: %d, applied: %d", len(pending), len(applied))
	return
}
