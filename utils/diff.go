package utils

import (
	"github.com/go-cinch/common/log"
	"github.com/r3labs/diff/v3"
)

// CompareDiff compare o(old struct) and n(new struct) to change, change must be pointer
func CompareDiff(o interface{}, n interface{}, change interface{}) {
	m1 := make(map[string]interface{}, 0)
	m2 := make(map[string]interface{}, 0)
	Struct2StructByJson(&m1, o)
	Struct2StructByJson(&m2, n)
	defer func() {
		err := recover()
		if err != nil {
			log.Warn("CompareDiff, pls check params: %v", err)
		}
	}()
	diff.Merge(m1, m2, change)
}
