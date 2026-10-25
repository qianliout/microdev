// Package rdb 提供 DAL 层使用的数据库连接封装。
package rdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"microdev/infra/consts"
)

// RDB 封装 *gorm.DB，是 DAL 层唯一的数据库入口。
type RDB struct {
	db *gorm.DB
}

func Open(driver, dsn string) (*RDB, error) {
	dialector, err := dialectorFor(driver, dsn)
	if err != nil {
		return nil, err
	}
	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	// SQLite 的内存库在最后一个连接关闭后会消失，且多连接易触发表锁，因此固定单连接。
	isSQLite := driver == consts.DBDriverSQLite
	if isSQLite {
		sqlDB.SetMaxOpenConns(1)
	}
	return &RDB{db: gdb}, nil
}

// Exec 执行一条不返回结果集的 SQL，用于建表等初始化语句。
func (r *RDB) Exec(ctx context.Context, sql string) error {
	return r.db.WithContext(ctx).Exec(sql).Error
}

// ExecAll 依次执行多条 SQL 并汇总所有错误。
func (r *RDB) ExecAll(ctx context.Context, stmts []string) error {
	errs := make([]error, 0, len(stmts))
	for _, stmt := range stmts {
		errs = append(errs, r.Exec(ctx, stmt))
	}
	return errors.Join(errs...)
}

func dialectorFor(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case consts.DBDriverSQLite:
		return sqlite.Open(dsn), nil
	case consts.DBDriverMySQL:
		return mysql.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q", driver)
	}
}

func (r *RDB) Get() *gorm.DB {
	return r.db
}

// Driver 返回当前连接使用的数据库驱动名。
func (r *RDB) Driver() string {
	return r.db.Dialector.Name()
}

func (r *RDB) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *RDB) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	return sqlDB.Close()
}
