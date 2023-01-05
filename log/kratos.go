package log

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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
	logDir = regexp.MustCompile(`kratos\.go`).ReplaceAllString(file, "")
	commonDir = regexp.MustCompile(`log.kratos\.go`).ReplaceAllString(file, "")
}

var _ Logger = (*kratosLog)(nil)

type kratosLog struct {
	log    *log.Helper
	ops    Options
	fields Fields
}

func Caller(ops Options) log.Valuer {
	return func(context.Context) interface{} {
		return caller(ops)
	}
}

func newKratosLog(ops *Options) *kratosLog {
	if ops.caller {
		ops.logger = log.With(
			ops.logger,
			CallerKey, Caller(*ops),
		)
	}
	helper := log.NewHelper(ops.logger)
	l := kratosLog{
		log:    helper,
		ops:    *ops,
		fields: make(Fields),
	}
	return &l
}

func (l *kratosLog) Options() Options {
	return l.ops
}

func (l *kratosLog) WithFields(fields Fields) Logger {
	ns := copyFields(fields)
	for k, v := range l.fields {
		ns[k] = v
	}
	ll := &kratosLog{
		log:    l.log,
		ops:    l.ops,
		fields: ns,
	}
	return ll
}

func (l *kratosLog) WithContext(ctx context.Context) Logger {
	ns := copyFields(l.fields)
	ll := &kratosLog{
		log:    l.log.WithContext(ctx),
		ops:    l.ops,
		fields: ns,
	}
	return ll
}

func (l *kratosLog) Log(level Level, args ...interface{}) {
	a := make([]interface{}, 0)
	a = append(a, l.ops.loggerMessageKey, fmt.Sprint(args...))
	for k, v := range l.fields {
		a = append(a, k, v)
	}
	l.log.Log(loggerToKratosLogLevel(level), a...)
}

func (l *kratosLog) Logf(level Level, format string, args ...interface{}) {
	a := make([]interface{}, 0)
	a = append(a, l.ops.loggerMessageKey, fmt.Sprintf(format, args...))
	for k, v := range l.fields {
		a = append(a, k, v)
	}
	l.log.Log(loggerToKratosLogLevel(level), a...)
}

func loggerToKratosLogLevel(level Level) log.Level {
	switch level {
	case TraceLevel, DebugLevel:
		return log.LevelDebug
	case InfoLevel:
		return log.LevelInfo
	case WarnLevel:
		return log.LevelWarn
	case ErrorLevel:
		return log.LevelError
	case FatalLevel:
		return log.LevelFatal
	default:
		return log.LevelInfo
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
		for _, item := range ops.callerSkip {
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
			return removeBaseDir(file+":"+strconv.FormatInt(int64(line), 10), ops)
		}
	}

	return ""
}

func removeBaseDir(s string, ops Options) string {
	sep := string(os.PathSeparator)
	if !ops.callerSource && strings.HasPrefix(s, commonDir) {
		s = strings.TrimPrefix(s, path.Dir(strings.TrimSuffix(commonDir, sep))+sep)
	}
	if ops.callerPrefix != "" && strings.HasPrefix(s, ops.callerPrefix) {
		s = strings.TrimPrefix(s, ops.callerPrefix)
	}
	arr := strings.Split(s, "@")
	if len(arr) == 2 {
		arr1 := strings.Split(arr[0], sep)
		arr2 := strings.Split(arr[1], sep)
		if ops.callerLevel > 0 {
			if ops.callerLevel < len(arr1) {
				arr1 = arr1[len(arr1)-ops.callerLevel:]
			}
		}
		if !ops.callerVersion {
			arr2 = arr2[1:]
		}
		s1 := strings.Join(arr1, sep)
		s2 := strings.Join(arr2, sep)
		if !ops.callerVersion {
			s = fmt.Sprintf("%s%s%s", s1, sep, s2)
		} else {
			s = fmt.Sprintf("%s@%s", s1, s2)
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
