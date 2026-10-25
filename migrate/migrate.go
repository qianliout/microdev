// Package migrate 内嵌各数据库驱动的建表 SQL，供 CLI 默认的内存库启动时自动建表。
// 外部数据库（如 MySQL）按项目约定由开发者手动执行 migrate/<driver>/ 下的脚本。
package migrate

import (
	_ "embed"
	"regexp"
	"slices"
	"strings"

	"microdev/infra/consts"
)

//go:embed sqlite/up.sql
var sqliteUp string

//go:embed mysql/up.sql
var mysqlUp string

var commentLine = regexp.MustCompile(`(?m)^[ \t]*--.*$`)

// Statements 返回指定驱动 up.sql 中的全部语句，已去掉注释行；驱动不支持时返回 nil。
func Statements(driver string) []string {
	switch driver {
	case consts.DBDriverSQLite:
		return splitStatements(sqliteUp)
	case consts.DBDriverMySQL:
		return splitStatements(mysqlUp)
	default:
		return nil
	}
}

func splitStatements(script string) []string {
	body := commentLine.ReplaceAllString(script, "")
	parts := strings.Split(body, ";")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return slices.DeleteFunc(parts, func(part string) bool { return part == "" })
}
