package i18n

import (
	"fmt"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"testing"
)

// //go:embed locales
// var fs embed.FS

func TestNew(t *testing.T) {
	i := New(
		WithFormat("yaml", yaml.Unmarshal),
	)
	// 1. add dir
	i.Add("./locales")

	// 2. add file
	// i.Add("./locales/en.yml")
	// i.Add("./locales/zh.yml")

	// 3. add embed fs
	// i.AddFs(fs)

	fmt.Println(i.T("common.hello"))
	fmt.Println(i.Select(language.Vietnamese).T("common.hello"))
}
