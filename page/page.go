package page

import (
	"gorm.io/gorm"
)

const (
	MinNum  uint64 = 1
	MinSize uint64 = 1
	Size    uint64 = 10
	MaxSize uint64 = 5000
)

// Page array data page info
type Page struct {
	Num     uint64 `json:"num"`     // current page
	Size    uint64 `json:"size"`    // page per count
	Total   int64  `json:"total"`   // all data count
	Disable bool   `json:"disable"` // disable pagination, query all data
}

// Scope returns a GORM scope function for pagination (without count)
// Usage: db.Scopes(page.Scope()).Find(&users)
func (page *Page) Scope() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page.Disable {
			return db
		}
		limit, offset := page.Limit()
		return db.Limit(limit).Offset(offset)
	}
}

// Limit calc limit/offset
func (page *Page) Limit() (int, int) {
	// handle negative total
	if page.Total < 0 {
		page.Total = 0
	}

	pageNum := page.Num
	pageSize := page.Size

	if pageNum < MinNum {
		pageNum = MinNum
	}
	if pageSize < MinSize || pageSize > MaxSize {
		pageSize = Size
	}

	offset := pageSize * (pageNum - 1)

	// if total exists, validate page number does not exceed max page
	if page.Total > 0 {
		maxPageNum := (uint64(page.Total) + pageSize - 1) / pageSize
		if pageNum > maxPageNum {
			pageNum = maxPageNum + 1
			offset = pageSize * maxPageNum
		}
	}

	page.Num = pageNum
	page.Size = pageSize
	if page.Disable {
		page.Size = uint64(page.Total)
	}

	return int(pageSize), int(offset)
}
