package main

// Two-language UI: Spanish when Windows (or $LANG) is Spanish, English
// otherwise. "lang = es|en" in skate3-mods.ini overrides it.

import (
	"os"
	"strings"
)

var langEN bool

func setLang(c config) {
	switch strings.ToLower(c["lang"]) {
	case "es":
		langEN = false
	case "en":
		langEN = true
	default:
		langEN = !systemSpanish()
	}
}

func envSpanish() bool {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return strings.HasPrefix(strings.ToLower(v), "es")
		}
	}
	return false
}

// T picks the Spanish or English text.
func T(es, en string) string {
	if langEN {
		return en
	}
	return es
}
