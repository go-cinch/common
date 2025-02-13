package i18n

import (
	"testing"

	"golang.org/x/text/language"
)

// //go:embed locales
// var fs embed.FS

func TestNew(t *testing.T) {
	i := New()
	// 1. add dir
	i.Add("./locales")

	// 2. add file
	// i.Add("./locales/en.yml")
	// i.Add("./locales/zh.yml")

	// 3. add embed fs
	// i.AddFs(fs)

	t.Log(i.T("common.hello"))
	t.Log(i.Select(language.Chinese).T("common.hello"))
}
