package utils

import (
	"github.com/go-cinch/common/log"
	"github.com/r3labs/diff/v3"
)

// CompareDiff compare o(old struct) and n(new struct) to change, change must be pointer
func CompareDiff(o interface{}, n interface{}, change interface{}) {
	m1 := make(map[string]interface{})
	m2 := make(map[string]interface{})
	Struct2StructByJSON(&m1, o)
	Struct2StructByJSON(&m2, n)
	m3 := make(map[string]interface{}, len(m1))
	m4 := make(map[string]interface{}, len(m2))
	for k, v := range m1 {
		m3[SnakeCase(k)] = v
	}
	for k, v := range m2 {
		m4[SnakeCase(k)] = v
	}
	defer func() {
		err := recover()
		if err != nil {
			log.Warn("CompareDiff, pls check params: %v", err)
		}
	}()
	_, err := diff.Merge(m3, m4, change)
	if err != nil {
		log.Warn("CompareDiff, pls check field type: %v", err)
	}
}
