package log

import "strings"

type Level uint32

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

var levelName = map[Level]string{
	PanicLevel: "panic",
	FatalLevel: "fatal",
	ErrorLevel: "error",
	WarnLevel:  "warn",
	InfoLevel:  "info",
	DebugLevel: "debug",
	TraceLevel: "trace",
}

var levelVal = map[string]Level{
	"panic": PanicLevel,
	"fatal": FatalLevel,
	"error": ErrorLevel,
	"warn":  WarnLevel,
	"info":  InfoLevel,
	"debug": DebugLevel,
	"trace": TraceLevel,
}

func NewLevel(name string) Level {
	name = strings.ToLower(name)
	if v, ok := levelVal[name]; ok {
		return v
	}
	// default info level
	return InfoLevel
}

func (l Level) Enabled(lvl Level) bool {
	return l >= lvl
}

func (l Level) String() string {
	return levelName[l]
}
