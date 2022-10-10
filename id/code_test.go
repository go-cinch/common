package id

import (
	"fmt"
	"testing"
)

func TestNewCode(t *testing.T) {
	var i uint64
	for i < 1000 {
		i++
		// use custom option
		item := NewCode(
			i,
			WithCodeChars([]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}),
			WithCodeN1(9),
			WithCodeN2(3),
			WithCodeL(5),
			WithCodeSalt(56789),
		)
		fmt.Println(item)
	}
	var j uint64
	for j < 1000 {
		j++
		// default option
		item := NewCode(j)
		fmt.Println(item)
	}
}
