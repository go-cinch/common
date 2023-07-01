package page

import (
	"context"
	"github.com/go-cinch/common/log"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

const (
	MinNum  uint64 = 1
	MinSize uint64 = 1
	Size    uint64 = 10
	MaxSize uint64 = 5000
)

// Page array data page info
type Page struct {
	ctx     context.Context
	Num     uint64 `json:"num"`     // current page
	Size    uint64 `json:"size"`    // page per count
	Total   int64  `json:"total"`   // all data count
	Disable bool   `json:"disable"` // disable pagination, query all data
	Count   bool   `json:"count"`   // not use 'SELECT count(*) FROM ...' before 'SELECT * FROM ...'
	Primary string `json:"primary"` // When there is a large amount of data, limit is optimized by specifying a field (the field is usually self incremented ID or indexed), which can improve the query efficiency (if it is not transmitted, it will not be optimized)
}

func (page *Page) WithContext(ctx context.Context) *Page {
	page.ctx = ctx
	return page
}

func (page *Page) Query(db *gorm.DB) (rp *Query) {
	rp = new(Query)
	rp.db = db
	if page.ctx == nil {
		page.ctx = context.Background()
	}
	rp.db.WithContext(page.ctx)
	rp.page = page
	return
}

// Limit calc limit/offset
func (page *Page) Limit() (int, int) {
	total := page.Total
	pageNum := page.Num
	pageSize := page.Size
	if page.Num < MinNum {
		pageNum = MinNum
	}
	if page.Size < MinSize || page.Size > MaxSize {
		pageSize = Size
	}

	// calc maxPageNum
	maxPageNum := uint64(total)/pageSize + 1
	if uint64(total)%pageSize == 0 {
		maxPageNum = uint64(total) / pageSize
	}
	if maxPageNum < MinNum {
		maxPageNum = MinNum
	}
	if total > 0 && pageNum > uint64(total) {
		pageNum = maxPageNum
	}

	limit := pageSize
	offset := limit * (pageNum - 1)
	// PageNum less than 1 is set as page 1 data
	if page.Num < 1 {
		offset = 0
	}

	// PageNum greater than maxPageNum is set as empty data: offset=last
	if total > 0 && page.Num > maxPageNum {
		pageNum = maxPageNum + 1
		offset = limit * maxPageNum
	}

	page.Num = pageNum
	page.Size = pageSize
	if page.Disable {
		page.Size = uint64(total)
	}
	// gorm v2 interface is int
	return int(limit), int(offset)
}

type Query struct {
	db   *gorm.DB
	page *Page
}

// Find exec gorm Find method with limit/offset
func (q *Query) Find(model interface{}) {
	db := q.db
	page := q.page
	if _, ok := db.Statement.Clauses["ORDER BY"]; !ok {
		db = db.Order(page.Primary)
	}
	ctx := page.ctx
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		log.WithContext(ctx).Warn("model must be a pointer")
		return
	}

	if !page.Disable {
		if !page.Count {
			db.Count(&page.Total)
		}
		if page.Total > 0 || page.Count {
			limit, offset := page.Limit()
			if page.Primary == "" {
				db.Limit(limit).Offset(offset).Find(model)
			} else {
				// parse model
				if db.Statement.Model != nil {
					err := db.Statement.Parse(db.Statement.Model)
					if err != nil {
						log.WithContext(ctx).WithError(err).Warn("parse model failed")
						return
					}
				}
				db.Joins(
					// add Primary index before join, improve query efficiency
					strings.Join([]string{
						"JOIN (?) AS `OFFSET_T` ON `",
						db.Statement.Table,
						"`.`",
						page.Primary,
						"` = `OFFSET_T`.`OFFSET_KEY`",
					}, ""),
					db.
						Session(&gorm.Session{}).
						Select(
							strings.Join([]string{
								"`",
								db.Statement.Table,
								"`.`",
								page.Primary,
								"` AS `OFFSET_KEY`",
							}, ""),
						).
						Limit(limit).
						Offset(offset),
				).Find(model)
			}
		}
	} else {
		// no pagination
		db.Find(model)
		page.Total = int64(rv.Elem().Len())
		page.Limit()
	}
	return
}

// Scan exec gorm Scan method with limit/offset
func (q *Query) Scan(model interface{}) {
	db := q.db
	page := q.page
	if _, ok := db.Statement.Clauses["ORDER BY"]; !ok {
		db = db.Order(page.Primary)
	}
	ctx := page.ctx
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		log.WithContext(ctx).Warn("model must be a pointer")
		return
	}

	if !page.Disable {
		if !page.Count {
			db.Count(&page.Total)
		}
		if page.Total > 0 || page.Count {
			limit, offset := page.Limit()
			if page.Primary == "" {
				db.Limit(limit).Offset(offset).Scan(model)
			} else {
				// parse model
				if db.Statement.Model != nil {
					err := db.Statement.Parse(db.Statement.Model)
					if err != nil {
						log.WithContext(ctx).WithError(err).Warn("parse model failed")
						return
					}
				}
				db.Joins(
					// add Primary index before join, improve query efficiency
					strings.Join([]string{
						"JOIN (?) AS `OFFSET_T` ON `",
						db.Statement.Table,
						"`.`",
						page.Primary,
						"` = `OFFSET_T`.`OFFSET_KEY`",
					}, ""),
					db.
						Session(&gorm.Session{}).
						Select(
							strings.Join([]string{
								"`",
								db.Statement.Table,
								"`.`",
								page.Primary,
								"` AS `OFFSET_KEY`",
							}, ""),
						).
						Limit(limit).
						Offset(offset),
				).Scan(model)
			}
		}
	} else {
		// no pagination
		db.Scan(model)
		page.Total = int64(rv.Elem().Len())
		page.Limit()
	}
	return
}
