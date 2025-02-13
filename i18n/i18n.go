package i18n

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type I18n struct {
	ops       Options
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	lang      language.Tag
}

func New(options ...func(*Options)) (rp *I18n) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	bundle := i18n.NewBundle(ops.language)
	localizer := i18n.NewLocalizer(bundle, ops.language.String())
	switch ops.format {
	case "toml":
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	case "json":
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	default:
		bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	}
	rp = &I18n{
		ops:       *ops,
		bundle:    bundle,
		localizer: localizer,
		lang:      ops.language,
	}
	for _, item := range ops.files {
		rp.Add(item)
	}
	rp.AddFs(ops.fs)
	return
}

// Select can change language
func (i I18n) Select(lang language.Tag) *I18n {
	if lang.String() == "und" {
		lang = i.ops.language
	}
	return &I18n{
		ops:       i.ops,
		bundle:    i.bundle,
		localizer: i18n.NewLocalizer(i.bundle, lang.String()),
		lang:      lang,
	}
}

// Language get current language
func (i I18n) Language() language.Tag {
	return i.lang
}

func (i I18n) T(id string) (rp string) {
	var err error
	rp, err = i.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: id,
		},
	})
	if err != nil || rp == "" {
		// use id as default message when unable to translate
		rp = id
	}
	return
}

func (i I18n) E(id string) error {
	return errors.Errorf(i.T(id))
}

// Add is add language file or dir(auto get language by filename)
func (i *I18n) Add(f string) {
	info, err := os.Stat(f)
	if err != nil {
		return
	}
	if info.IsDir() {
		filepath.Walk(f, func(path string, fi os.FileInfo, _ error) (err error) {
			if !fi.IsDir() {
				i.bundle.LoadMessageFile(path)
			}
			return
		})
	} else {
		i.bundle.LoadMessageFile(f)
	}
}

// AddFs is add language embed files
func (i *I18n) AddFs(fs embed.FS) {
	files := readFs(fs, ".")
	for _, name := range files {
		i.bundle.LoadMessageFileFS(fs, name)
	}
}

func readFs(fs embed.FS, dir string) (rp []string) {
	rp = make([]string, 0)
	dirs, err := fs.ReadDir(dir)
	if err != nil {
		return
	}
	for _, item := range dirs {
		name := filepath.ToSlash(filepath.Join(dir, item.Name()))
		if dir == "." {
			name = item.Name()
		}
		if item.IsDir() {
			rp = append(rp, readFs(fs, name)...)
		} else {
			rp = append(rp, name)
		}
	}
	return
}
