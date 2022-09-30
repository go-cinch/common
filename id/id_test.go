package id

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	var i uint64
	for i < 1000 {
		i++
		// use custom option
		item := New(
			i,
			WithChars([]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}),
			WithN1(9),
			WithN2(3),
			WithL(5),
			WithSalt(56789),
		)
		fmt.Println(item)
	}
	var j uint64
	for j < 1000 {
		j++
		// default option
		item := New(j)
		fmt.Println(item)
	}
}
