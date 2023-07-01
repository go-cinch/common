package utils

import "testing"

func TestContains(t *testing.T) {
	type args[T comparable] struct {
		arr  []T
		item T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}
	intTests := []testCase[int]{
		{
			name: "int1",
			args: args[int]{
				arr:  []int{1, 2, 3},
				item: 1,
			},
			want: true,
		},
		{
			name: "int2",
			args: args[int]{
				arr:  []int{1, 2, 3},
				item: 5,
			},
			want: false,
		},
	}
	for _, tt := range intTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.arr, tt.args.item); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
	stringTests := []testCase[string]{
		{
			name: "string1",
			args: args[string]{
				arr:  []string{"1", "2", "3"},
				item: "1",
			},
			want: true,
		},
		{
			name: "string2",
			args: args[string]{
				arr:  []string{"1", "2", "3"},
				item: "5",
			},
			want: false,
		},
	}
	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.arr, tt.args.item); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
	uintTests := []testCase[uint8]{
		{
			name: "string1",
			args: args[uint8]{
				arr:  []uint8{1, 2, 3},
				item: 1,
			},
			want: true,
		},
		{
			name: "string2",
			args: args[uint8]{
				arr:  []uint8{1, 2, 3},
				item: 5,
			},
			want: false,
		},
	}
	for _, tt := range uintTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.arr, tt.args.item); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
