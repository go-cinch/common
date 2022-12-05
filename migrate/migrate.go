package migrate

import (
	"database/sql"
	"fmt"
	"github.com/go-cinch/common/log"
	m "github.com/go-sql-driver/mysql"
	migrate "github.com/rubenv/sql-migrate"
	"os"
	"strings"
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
		} else {
			log.
				WithContext(ops.ctx).
				WithFields(log.Fields{
					"migrate.lock": ops.lockName,
				}).
				Info("cannot acquire advisory lock, retrying...")
		}
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
		arr := strings.Split(rollback, "; ")
		for i, item := range arr {
			_, err = db.ExecContext(ops.ctx, item)
			if err != nil {
				log.
					WithContext(ops.ctx).
					WithError(err).
					WithField("rollback.sql", rollback).
					WithField(fmt.Sprintf("rollback.%d", i+1), item).
					Error("exec rollback failed")
				return
			}
		}
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
	var cfg *m.Config
	cfg, err = m.ParseDSN(ops.uri)
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("invalid uri")
		return
	}
	dbname := cfg.DBName
	cfg.DBName = ""
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			Error("create database failed")
	}
	return
}

func acquireLock(ops *Options, db *sql.DB) (f bool, err error) {
	// GET_LOCK will be blocked if another session already acquired the lock
	// timeout 5s
	q := fmt.Sprintf("SELECT GET_LOCK('%v', 5)", ops.lockName)
	err = db.QueryRow(q).Scan(&f)

	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			WithFields(log.Fields{
				"migrate.lock": ops.lockName,
			}).
			Error("acquire advisory lock for migration failed")
		return
	}

	log.
		WithContext(ops.ctx).
		WithFields(log.Fields{
			"migrate.lock": ops.lockName,
		}).
		Info("acquire advisory lock: %v", f)
	return
}

func releaseLock(ops *Options, db *sql.DB) (err error) {
	q := fmt.Sprintf("SELECT RELEASE_LOCK('%v')", ops.lockName)
	_, err = db.Exec(q)

	if err != nil {
		log.
			WithContext(ops.ctx).
			WithError(err).
			WithFields(log.Fields{
				"migrate.lock": ops.lockName,
			}).
			Error("release advisory lock for migration failed")
		return err
	}

	log.
		WithContext(ops.ctx).
		WithFields(log.Fields{
			"migrate.lock": ops.lockName,
		}).
		Info("release advisory lock success")
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
