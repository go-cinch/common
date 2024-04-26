package caller

import (
	"regexp"
	"testing"
)

func TestCaller(t *testing.T) {
	arr := []string{
		"github.com/go-cinch/common/log@v1.0.7-0.20240426023506-ea8f2551f73f/caller/caller.go",
		"github.com/go-cinch/common/log@v1.0.8/caller/caller.go",
		"github.com/go-cinch/common/log@v1.0/caller/caller.go",
		"github.com/go-cinch/common/log@/caller/caller.go",
		"github.com/go-cinch/common/log/caller/caller.go",
	}

	target := "github.com/go-cinch/common/log/caller/caller.go"

	pattern := `@v\d+(\.\d+)+(-\d+\.\d+-[a-f0-9]+)?|@`

	re := regexp.MustCompile(pattern)

	for _, item := range arr {
		result := re.ReplaceAllString(item, "")
		if result != target {
			t.Errorf("expect %s, got %s", target, result)
			return
		}
	}
}
