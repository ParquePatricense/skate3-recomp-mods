package main

// Seven-language UI: English, Spanish, French, Italian, German, Chinese and
// Japanese. The system language picks it (Windows' display language, $LANG
// elsewhere); "lang = <code>" in skate3-mods.ini overrides it.
//
// Texts are written as T(spanish, english); the other languages come from
// i18n/go.json, keyed by the English text.

import (
	_ "embed"
	"encoding/json"
	"os"
	"strings"
)

var langs = []struct{ code, name string }{
	{"en", "English"}, {"es", "Español"}, {"fr", "Français"}, {"it", "Italiano"},
	{"de", "Deutsch"}, {"zh", "中文"}, {"ja", "日本語"},
}

var (
	lang   = "en"
	langEN = true // English or a language without a Spanish text: the English one is the base

	//go:embed i18n/go.json
	goI18nJSON []byte
	goI18n     map[string]map[string]string // English text -> language -> text
)

func validLang(code string) bool {
	for _, l := range langs {
		if l.code == code {
			return true
		}
	}
	return false
}

func setLang(c config) {
	l := strings.ToLower(c["lang"])
	if !validLang(l) {
		l = systemLang()
	}
	lang, langEN = l, l != "es"
}

// langFromCode turns a locale such as "es_AR.UTF-8" or "zh-Hans" into one of
// the seven languages (English when it is none of them).
func langFromCode(s string) string {
	s = strings.ToLower(s)
	if len(s) >= 2 && validLang(s[:2]) {
		return s[:2]
	}
	return "en"
}

func envLang() string {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return langFromCode(v)
		}
	}
	return "en"
}

func systemSpanish() bool { return systemLang() == "es" }

// T picks the text in the current language.
func T(es, en string) string {
	switch lang {
	case "es":
		return es
	case "en":
		return en
	}
	if goI18n == nil {
		goI18n = map[string]map[string]string{}
		json.Unmarshal(goI18nJSON, &goI18n)
	}
	if s := goI18n[en][lang]; s != "" {
		return s
	}
	return en
}
