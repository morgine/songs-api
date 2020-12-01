package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Query struct {
	Query string
	Args  []interface{}
}

func (q Query) AppendCondition(db *gorm.DB) *gorm.DB {
	return db.Where(q.Query, q.Args...)
}

func Where(query string, args ...interface{}) Query {
	return Query{Query: query, Args: args}
}

type ConditionFunc func(db *gorm.DB) *gorm.DB

func (c ConditionFunc) AppendCondition(db *gorm.DB) *gorm.DB {
	return c(db)
}

type Pagination struct {
	Limit  *int
	Offset *int
}

func (p Pagination) AppendCondition(db *gorm.DB) *gorm.DB {
	if p.Limit == nil {
		p.Limit = newInt(10)
	}
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
	}
	if p.Offset != nil {
		db = db.Offset(*p.Offset)
	}
	return db
}

func newInt(i int) *int {
	return &i
}

type OrderBy struct {
	Column  string // 排序字段
	Desc    bool   // 逆序
	Reorder bool   // 重排序
}

func (o OrderBy) AppendCondition(db *gorm.DB) *gorm.DB {
	if o.Column == "" {
		o.Column = "id"
	}
	return db.Order(clause.OrderByColumn{Column: clause.Column{Name: o.Column}, Desc: o.Desc})
}

type Condition interface {
	AppendCondition(db *gorm.DB) *gorm.DB
}

func HandleConditions(db *gorm.DB, c ...Condition) *gorm.DB {
	for _, condition := range c {
		db = condition.AppendCondition(db)
	}
	return db
}
