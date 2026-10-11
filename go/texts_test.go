package main

import (
	"bytes"
	"os"
	"testing"
)

func TestApplyRules(t *testing.T) {
	r := loadRules()
	cases := map[string]string{
		"Haz un kickflip y presiona <B:A_hud>":         "Hacé un kickflip y presioná <B:A_hud>",
		"¿Quieres jugar? Puedes probar, ¿vale?":        "¿Querés jugar? Podés probar, ¿dale?",
		"Prueba en curso":                              "Prueba en curso",
		"Prueba esta línea, colega.":                   "Probá esta línea, loco.",
		"Más vale que aprietes":                        "Más vale que aprietes",
		"Lánzate desde el techo. Fíjate en los coches": "Lanzate desde el techo. Fijate en los autos",
		"PUEDES":                 "PODÉS",
		"Suma {0:points} puntos": "Sumá {0:points} puntos",
	}
	for in, want := range cases {
		if got := r.apply(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}

// S3_LANG_BIN=path/to/LANGUAGE_Spanish_Global_skate3ng.BIN runs it on real data.
func TestRioplatensizeFile(t *testing.T) {
	p := os.Getenv("S3_LANG_BIN")
	if p == "" {
		t.Skip("S3_LANG_BIN not set")
	}
	bin, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	out, changed, err := rioplatensize(bin)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("changed %d strings, %d -> %d bytes", changed, len(bin), len(out))
	if !bytes.Equal(out[8:0x24], bin[8:0x24]) {
		t.Error("header changed")
	}
	if dst := os.Getenv("S3_LANG_OUT"); dst != "" {
		os.WriteFile(dst, out, 0o644)
	}
}
