package utils

import (
	"encoding/json"
	"github.com/go-cinch/common/log"
)

func Struct2Json(to *string, from interface{}) {
	str, err := json.Marshal(from)
	if err != nil {
		log.Warn("Struct2Json, can not convert: %v", err)
	}
	if to == nil {
		log.Warn("Struct2Json, target is empty")
		return
	}
	*to = string(str)
}

func Json2Struct(to interface{}, from string) {
	err := json.Unmarshal([]byte(from), to)
	if err != nil {
		log.Warn("Json2Struct, can not convert: %v", err)
	}
}

func Struct2StructByJson(to interface{}, from interface{}) {
	var str string
	Struct2Json(&str, from)
	Json2Struct(to, str)
}
