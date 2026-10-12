package main

import "testing"

func TestTranslations(t *testing.T) {
	defer setLang(config{"lang": "es"})
	for _, l := range langs {
		setLang(config{"lang": l.code})
		got := T("Listo.", "Done.")
		if got == "" || (l.code != "en" && l.code != "es" && got == "Done.") {
			t.Errorf("%s: %q", l.code, got)
		}
	}
	setLang(config{"lang": "ja"})
	if !langEN || lang != "ja" {
		t.Fatal(lang)
	}
	if langFromCode("zh_CN.UTF-8") != "zh" || langFromCode("pt_BR") != "pt" || langFromCode("ko_KR") != "en" {
		t.Fatal("langFromCode")
	}
}
