package page

import (
	"testing"
)

func TestLimit(t *testing.T) {
	tests := []struct {
		name       string
		page       Page
		wantLimit  int
		wantOffset int
		wantNum    int64
		wantSize   int64
		wantTotal  int64
	}{
		{
			name:       "normal case: page 1, size 10, total 100",
			page:       Page{Num: 1, Size: 10, Total: 100},
			wantLimit:  10,
			wantOffset: 0,
			wantNum:    1,
			wantSize:   10,
			wantTotal:  100,
		},
		{
			name:       "normal case: page 2, size 10, total 100",
			page:       Page{Num: 2, Size: 10, Total: 100},
			wantLimit:  10,
			wantOffset: 10,
			wantNum:    2,
			wantSize:   10,
			wantTotal:  100,
		},
		{
			name:       "normal case: page 5, size 20, total 100",
			page:       Page{Num: 5, Size: 20, Total: 100},
			wantLimit:  20,
			wantOffset: 80,
			wantNum:    5,
			wantSize:   20,
			wantTotal:  100,
		},
		{
			name:       "page num is 0, should be set to 1",
			page:       Page{Num: 0, Size: 10, Total: 100},
			wantLimit:  10,
			wantOffset: 0,
			wantNum:    1,
			wantSize:   10,
			wantTotal:  100,
		},
		{
			name:       "page size is 0, should use default size 10",
			page:       Page{Num: 1, Size: 0, Total: 100},
			wantLimit:  10,
			wantOffset: 0,
			wantNum:    1,
			wantSize:   10,
			wantTotal:  100,
		},
		{
			name:       "page size exceeds max, should use default size 10",
			page:       Page{Num: 1, Size: 6000, Total: 100},
			wantLimit:  10,
			wantOffset: 0,
			wantNum:    1,
			wantSize:   10,
			wantTotal:  100,
		},
		{
			name:       "page num exceeds max page, should return empty offset",
			page:       Page{Num: 20, Size: 10, Total: 100},
			wantLimit:  10,
			wantOffset: 100,
			wantNum:    11,
			wantSize:   10,
			wantTotal:  100,
		},
		{
			name:       "total is 0",
			page:       Page{Num: 1, Size: 10, Total: 0},
			wantLimit:  10,
			wantOffset: 0,
			wantNum:    1,
			wantSize:   10,
			wantTotal:  0,
		},
		{
			name:       "negative total should be set to 0",
			page:       Page{Num: 1, Size: 10, Total: -5},
			wantLimit:  10,
			wantOffset: 0,
			wantNum:    1,
			wantSize:   10,
			wantTotal:  0,
		},
		{
			name:       "last page with partial records: total 95, size 10, page 10",
			page:       Page{Num: 10, Size: 10, Total: 95},
			wantLimit:  10,
			wantOffset: 90,
			wantNum:    10,
			wantSize:   10,
			wantTotal:  95,
		},
		{
			name:       "exact page boundary: total 100, size 10, page 10",
			page:       Page{Num: 10, Size: 10, Total: 100},
			wantLimit:  10,
			wantOffset: 90,
			wantNum:    10,
			wantSize:   10,
			wantTotal:  100,
		},
		{
			name:       "disable pagination sets size to total",
			page:       Page{Num: 1, Size: 10, Total: 50, Disable: true},
			wantLimit:  10,
			wantOffset: 0,
			wantNum:    1,
			wantSize:   50,
			wantTotal:  50,
		},

		{
			name:       "page 1 with size 1",
			page:       Page{Num: 1, Size: 1, Total: 10},
			wantLimit:  1,
			wantOffset: 0,
			wantNum:    1,
			wantSize:   1,
			wantTotal:  10,
		},
		{
			name:       "max size boundary",
			page:       Page{Num: 1, Size: MaxSize, Total: 10000},
			wantLimit:  int(MaxSize),
			wantOffset: 0,
			wantNum:    1,
			wantSize:   MaxSize,
			wantTotal:  10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLimit, gotOffset := tt.page.Limit()

			if gotLimit != tt.wantLimit {
				t.Errorf("Limit() gotLimit = %v, want %v", gotLimit, tt.wantLimit)
			}
			if gotOffset != tt.wantOffset {
				t.Errorf("Limit() gotOffset = %v, want %v", gotOffset, tt.wantOffset)
			}
			if tt.page.Num != tt.wantNum {
				t.Errorf("Limit() page.Num = %v, want %v", tt.page.Num, tt.wantNum)
			}
			if tt.page.Size != tt.wantSize {
				t.Errorf("Limit() page.Size = %v, want %v", tt.page.Size, tt.wantSize)
			}
			if tt.page.Total != tt.wantTotal {
				t.Errorf("Limit() page.Total = %v, want %v", tt.page.Total, tt.wantTotal)
			}
		})
	}
}

func TestScope(t *testing.T) {
	// Test that Scope returns a non-nil function
	p := &Page{Num: 1, Size: 10}
	scopeFunc := p.Scope()
	if scopeFunc == nil {
		t.Error("Scope() returned nil")
	}
}
