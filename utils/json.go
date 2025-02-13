package utils

import (
	"encoding/json"

	"github.com/go-cinch/common/log"
)

func Struct2JSON(from interface{}) string {
	str, err := json.Marshal(from)
	if err != nil {
		log.Warn("Struct2JSON, can not convert: %v", err)
	}
	return string(str)
}

func JSON2Struct(to interface{}, from string) {
	err := json.Unmarshal([]byte(from), to)
	if err != nil {
		log.Warn("JSON2Struct, can not convert: %v", err)
	}
}

func Struct2StructByJSON(to interface{}, from interface{}) {
	str := Struct2JSON(from)
	JSON2Struct(to, str)
}
