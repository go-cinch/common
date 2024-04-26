package caller

import (
	"context"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

var (
	logDir      = ""
	commonDir   = ""
	notSkipDirs = []string{
		"transport/grpc/server",
		"transport/http/server",
		"middleware/logging/logging",
	}
)

func init() {
	// get runtime root
	_, file, _, _ := runtime.Caller(0)
	file = regexp.MustCompile(`@v\d+(\.\d+)+(-\d+\.\d+-[a-f0-9]+)?|@`).ReplaceAllString(file, "")
	logDir = regexp.MustCompile(`caller/caller\.go`).ReplaceAllString(file, "")
	commonDir = regexp.MustCompile(`log/caller/caller\.go`).ReplaceAllString(file, "")
}

func Caller(options ...func(*Options)) log.Valuer {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(context.Context) interface{} {
		return caller(*ops)
	}
}

func caller(ops Options) string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 0; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if strings.Contains(file, commonDir) {
			continue
		}
		var skip bool
		for _, item := range ops.skips {
			if strings.Contains(file, item) {
				skip = true
				break
			}
		}
		if strings.Contains(file, "go-kratos") && containsString(notSkipDirs, file) {
			skip = false
		}
		if skip {
			continue
		}
		if ok && (!strings.HasPrefix(file, logDir) || strings.HasSuffix(file, "_test.go")) && !strings.Contains(file, "src/runtime") {
			return removeBaseDir(strings.Join([]string{file, strconv.Itoa(line)}, ":"), ops)
		}
	}

	return ""
}

func removeBaseDir(s string, ops Options) string {
	sep := string(os.PathSeparator)
	if !ops.source && strings.HasPrefix(s, commonDir) {
		s = strings.TrimPrefix(s, strings.Join([]string{path.Dir(strings.TrimSuffix(commonDir, sep)), sep}, ""))
	}
	if ops.prefix != "" && strings.HasPrefix(s, ops.prefix) {
		s = strings.TrimPrefix(s, ops.prefix)
	}
	arr := strings.Split(s, "@")
	if len(arr) == 2 {
		arr1 := strings.Split(arr[0], sep)
		arr2 := strings.Split(arr[1], sep)
		if ops.level > 0 {
			if ops.level < len(arr1) {
				arr1 = arr1[len(arr1)-ops.level:]
			}
		}
		if !ops.version {
			arr2 = arr2[1:]
		}
		s1 := strings.Join(arr1, sep)
		s2 := strings.Join(arr2, sep)
		if !ops.version {
			s = strings.Join([]string{s1, s2}, sep)
		} else {
			s = strings.Join([]string{s1, s2}, "@")
		}
	}
	return s
}

func containsString(s []string, v string) bool {
	for _, vv := range s {
		if strings.Contains(v, vv) {
			return true
		}
	}
	return false
}
