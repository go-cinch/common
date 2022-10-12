package page

import (
	"context"
	"fmt"
	"github.com/go-cinch/common/log"
	"gorm.io/gorm"
	"reflect"
)

const (
	MinNum  uint64 = 1
	MinSize uint64 = 1
	Size    uint64 = 10
	MaxSize uint64 = 5000
)

// Page array data page info
type Page struct {
	ctx      context.Context
	PageNum  uint64 `json:"pageNum"`  // current page
	PageSize uint64 `json:"pageSize"` // page per count
	Total    int64  `json:"total"`    // all data count
	Disable  bool   `json:"disable"`  // disable pagination, query all data
	Count    bool   `json:"count"`    // not use 'SELECT count(*) FROM ...' before 'SELECT * FROM ...'
	Primary  string `json:"primary"`  // When there is a large amount of data, limit is optimized by specifying a field (the field is usually self incremented ID or indexed), which can improve the query efficiency (if it is not transmitted, it will not be optimized)
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
	rp.page = page
	return
}

// Limit calc limit/offset
func (page *Page) Limit() (int, int) {
	total := page.Total
	pageNum := page.PageNum
	pageSize := page.PageSize
	if page.PageNum < MinNum {
		pageNum = MinNum
	}
	if page.PageSize < MinSize || page.PageSize > MaxSize {
		pageSize = Size
	}

	// calc maxPageNum
	maxPageNum := uint64(total)/pageSize + 1
	if uint64(total)%pageSize == 0 {
		maxPageNum = uint64(total) / pageSize
	}
	// maxPageNum must be greater than 0
	if maxPageNum < MinNum {
		maxPageNum = MinNum
	}
	// pageNum must be less than or equal to total
	if total > 0 && pageNum > uint64(total) {
		pageNum = maxPageNum
	}

	limit := pageSize
	offset := limit * (pageNum - 1)
	// PageNum less than 1 is set as page 1 data
	if page.PageNum < 1 {
		offset = 0
	}

	// PageNum greater than maxPageNum is set as empty data: offset=last
	if total > 0 && page.PageNum > maxPageNum {
		pageNum = maxPageNum + 1
		offset = limit * maxPageNum
	}

	page.PageNum = pageNum
	page.PageSize = pageSize
	if page.Disable {
		page.PageSize = uint64(total)
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
					fmt.Sprintf(
						"JOIN (?) AS `OFFSET_T` ON `%s`.`%s` = `OFFSET_T`.`OFFSET_KEY`",
						db.Statement.Table,
						page.Primary,
					),
					db.
						Session(&gorm.Session{}).
						Select(
							fmt.Sprintf("`%s`.`%s` AS `OFFSET_KEY`", db.Statement.Table, page.Primary),
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
					fmt.Sprintf(
						"JOIN (?) AS `OFFSET_T` ON `%s`.`%s` = `OFFSET_T`.`OFFSET_KEY`",
						db.Statement.Table,
						page.Primary,
					),
					db.
						Session(&gorm.Session{}).
						Select(
							fmt.Sprintf("`%s`.`%s` AS `OFFSET_KEY`", db.Statement.Table, page.Primary),
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
