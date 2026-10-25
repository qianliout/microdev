// Package filter 提供 DAL 查询通用的分页与排序条件。
package filter

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	SortAsc  = "asc"
	SortDesc = "desc"
)

type Filter struct {
	SortBy    string `json:"sortBy"`
	SortFiled string `json:"sortFiled"`

	// 多列或显式排序。
	OrderByColumns []clause.OrderByColumn `json:"-"`

	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

func (vi *Filter) DeepCopy() *Filter {
	if vi == nil {
		return &Filter{}
	}
	cp := *vi
	cp.OrderByColumns = append([]clause.OrderByColumn(nil), vi.OrderByColumns...)
	return &cp
}

// AddFilter 把分页与排序条件追加到查询上，filter 为 nil 时不加任何限制。
func AddFilter(db *gorm.DB, filter *Filter) *gorm.DB {
	if filter == nil {
		return db
	}
	if filter.Offset >= 0 && filter.Limit > 0 {
		db = db.Offset(int(filter.Offset)).Limit(int(filter.Limit))
	}
	if filter.SortFiled != "" && filter.SortBy != "" {
		desc := strings.ToLower(filter.SortBy) == SortDesc
		db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: filter.SortFiled}, Desc: desc})
	}
	for _, column := range filter.OrderByColumns {
		db = db.Order(column)
	}
	return db
}
