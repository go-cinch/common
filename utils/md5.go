package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// StructMd5 get struct md5 string by json
func StructMd5(o interface{}) string {
	str := Struct2Json(o)
	h := md5.New()
	h.Write([]byte(str))
	data := h.Sum([]byte(nil))
	return hex.EncodeToString(data)
}
