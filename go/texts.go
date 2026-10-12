// Rioplatense texts module: adds "Español ARG" to the game's language
// screen (in Dutch's slot, inside game/data/big/miscboot.big), built from the
// Spanish texts with voseo and Argentine vocabulary (rules in rioplatense.txt).
// Only rules ship with the mod; the game text comes from the user's copy.
package main

import (
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	textsArchive = "miscboot.big"
	textsEntry   = "LANGUAGE_Spanish_Global_skate3ng.BIN"
	textsBackup  = ".original"
	textsState   = ".rioplatense"
	langNameHash = 0x802244c6 // the "Nederlands" entry of the language list
	langNameAR   = "Español ARG"
)

//go:embed rioplatense.txt
var rulesText string

type textRules struct {
	words   map[string]string // anywhere
	starts  map[string]string // sentence start only
	phrases [][2]string       // literal, in file order, before the word pass
	fixes   [][2]string       // literal, after the word pass
}

func loadRules() textRules {
	r := textRules{words: map[string]string{}, starts: map[string]string{}}
	for _, ln := range strings.Split(rulesText, "\n") {
		f := strings.Fields(ln)
		if len(f) != 3 || strings.HasPrefix(f[0], "#") {
			continue
		}
		from, to := strings.ReplaceAll(f[1], "_", " "), strings.ReplaceAll(f[2], "_", " ")
		switch f[0] {
		case "w":
			r.words[from] = to
		case "s":
			r.starts[from] = to
		case "p":
			r.phrases = append(r.phrases, [2]string{from, to})
		case "q":
			r.fixes = append(r.fixes, [2]string{from, to})
		}
	}
	return r
}

func isLetter(c rune) bool { return unicode.IsLetter(c) }

// matchCase gives repl the capitalization of orig (Title, UPPER or lower).
func matchCase(orig, repl string) string {
	o := []rune(orig)
	if len(o) > 1 && strings.ToUpper(orig) == orig {
		return strings.ToUpper(repl)
	}
	if unicode.IsUpper(o[0]) {
		r := []rune(repl)
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}
	return repl
}

// imperativeLead lists words after which an ambiguous verb form is read as
// an order ("y luego presiona", "Eh, mira", "Tú supera").
var imperativeLead = map[string]bool{"y": true, "luego": true, "ahora": true, "después": true,
	"eh": true, "bien": true, "pues": true, "tú": true, "vos": true, "mejor": true}

func atSentenceStart(s []rune, i int) bool {
	j := i - 1
	for j >= 0 && (s[j] == ' ' || s[j] == '"' || s[j] == '«') {
		j--
	}
	if j < 0 {
		return true
	}
	if strings.ContainsRune(".!?:^\n¡¿]", s[j]) {
		return true
	}
	if s[j] == ',' { // "Eh, mira" / "Bien, haz"
		j--
	}
	k := j
	for k >= 0 && isLetter(s[k]) {
		k--
	}
	return k < j && imperativeLead[strings.ToLower(string(s[k+1:j+1]))]
}

// nextWord returns the word after position i (lowercased), or "" at the end.
func nextWord(s []rune, i int) string {
	for i < len(s) && s[i] == ' ' {
		i++
	}
	if i >= len(s) {
		return ""
	}
	j := i
	for j < len(s) && isLetter(s[j]) {
		j++
	}
	if j == i {
		return string(s[i])
	}
	return strings.ToLower(string(s[i:j]))
}

func (r textRules) apply(in string) string {
	s := in
	for _, p := range r.phrases {
		s = strings.ReplaceAll(s, p[0], p[1])
	}
	src := []rune(s)
	var out strings.Builder
	for i := 0; i < len(src); {
		c := src[i]
		if c == '<' || c == '{' { // button tags / format fields stay as they are
			end := map[rune]rune{'<': '>', '{': '}'}[c]
			j := i
			for j < len(src) && src[j] != end {
				j++
			}
			if j < len(src) {
				j++
			}
			out.WriteString(string(src[i:j]))
			i = j
			continue
		}
		if !isLetter(c) {
			out.WriteRune(c)
			i++
			continue
		}
		j := i
		for j < len(src) && isLetter(src[j]) {
			j++
		}
		w := string(src[i:j])
		lw := strings.ToLower(w)
		if to, ok := r.words[lw]; ok {
			w = matchCase(w, to)
		} else if to, ok := r.starts[lw]; ok && atSentenceStart(src, i) {
			switch nw := nextWord(src, j); {
			case nw == "" || nw == "de" || nw == "del" || nw == "en":
			case lw == "vale" && nw == "la":
			default:
				w = matchCase(w, to)
			}
		}
		out.WriteString(w)
		i = j
	}
	res := out.String()
	for _, p := range r.fixes {
		res = strings.ReplaceAll(res, p[0], p[1])
	}
	return res
}

func latin1ToString(b []byte) string {
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func stringToLatin1(s string) ([]byte, error) {
	out := make([]byte, 0, len(s))
	for _, c := range s {
		if c > 0xFF {
			return nil, fmt.Errorf("carácter fuera de latin-1: %q", c)
		}
		out = append(out, byte(c))
	}
	return out, nil
}

// rewriteTable rewrites every string of a LANGUAGE_*_Global .BIN table
// through fn. Layout (little endian): +0 magic, +4 size-8, +8 count,
// +12 table offset (from +8), +16 strings offset (from +8), +20 name;
// table = count x (hash, offset into strings); strings are latin-1, NUL
// terminated.
func rewriteTable(bin []byte, fn func(hash uint32, s string) string) ([]byte, int, error) {
	bad := errors.New(T("tabla de textos con formato inesperado", "unexpected text table format"))
	if len(bin) < 0x24 || int(binary.LittleEndian.Uint32(bin[4:]))+8 != len(bin) {
		return nil, 0, bad
	}
	n := int(binary.LittleEndian.Uint32(bin[8:]))
	tab := 8 + int(binary.LittleEndian.Uint32(bin[12:]))
	base := 8 + int(binary.LittleEndian.Uint32(bin[16:]))
	if tab+n*8 > base || base > len(bin) {
		return nil, 0, bad
	}
	out := append([]byte(nil), bin[:base]...)
	newOff := map[uint32]uint32{}
	changed := 0
	for i := 0; i < n; i++ {
		e := tab + i*8
		hash, off := binary.LittleEndian.Uint32(bin[e:]), binary.LittleEndian.Uint32(bin[e+4:])
		if no, ok := newOff[off]; ok {
			binary.LittleEndian.PutUint32(out[e+4:], no)
			continue
		}
		start := base + int(off)
		if start >= len(bin) {
			return nil, 0, bad
		}
		end := start
		for end < len(bin) && bin[end] != 0 {
			end++
		}
		str, err := stringToLatin1(fn(hash, latin1ToString(bin[start:end])))
		if err != nil {
			return nil, 0, err
		}
		if string(str) != string(bin[start:end]) {
			changed++
		}
		no := uint32(len(out) - base)
		newOff[off] = no
		binary.LittleEndian.PutUint32(out[e+4:], no)
		out = append(out, str...)
		out = append(out, 0)
	}
	for len(out)%4 != 0 {
		out = append(out, 0)
	}
	binary.LittleEndian.PutUint32(out[4:], uint32(len(out)-8))
	return out, changed, nil
}

// rioplatensize turns a Spanish table into the Rioplatense one.
func rioplatensize(bin []byte) ([]byte, int, error) {
	rules := loadRules()
	return rewriteTable(bin, func(h uint32, s string) string {
		if h == langNameHash {
			return langNameAR
		}
		return rules.apply(s)
	})
}

func textsArchivePath(game string) (string, error) {
	icons, err := findArchive(game)
	if err != nil {
		return "", err
	}
	p := filepath.Join(filepath.Dir(icons), textsArchive)
	if !exists(p) {
		return "", fmt.Errorf(T("no encontré %s", "could not find %s"), p)
	}
	return p, nil
}

// installTexts turns the Dutch slot of the language screen into
// "Español ARG": Dutch gets the Rioplatense version of the Spanish texts
// (plus Spanish's glyph histogram and locale file), and every language's
// list shows "Español ARG" instead of "Nederlands". Spanish stays as is.
func installTexts(root, game string) error {
	archive, err := textsArchivePath(game)
	if err != nil {
		return err
	}
	backup := archive + textsBackup
	if !exists(backup) {
		fmt.Printf(T("Guardando copia del %s original...\n", "Backing up the original %s...\n"), textsArchive)
		if err := copyFile(archive, backup); err != nil {
			return err
		}
	}
	orig, err := os.ReadFile(backup)
	if err != nil {
		return err
	}
	big, err := parseBig(orig)
	if err != nil {
		return err
	}
	get := func(name string) (int, []byte, error) {
		i, err := big.find(name)
		if err != nil {
			return 0, nil, err
		}
		d, err := big.read(i)
		return i, d, err
	}
	fmt.Println(T("Armando el idioma \"Español ARG\"...", "Building the \"Español ARG\" language..."))
	_, es, err := get(textsEntry)
	if err != nil {
		return err
	}
	ar, changed, err := rioplatensize(es)
	if err != nil {
		return err
	}
	copies := [][2]string{
		{"LANGUAGE_Spanish_HISTOGRAM_skate3ng.BIN", "LANGUAGE_Dutch_HISTOGRAM_skate3ng.BIN"},
		{"LOCALE_SPANISH.loc", "LOCALE_DUTCH.loc"},
	}
	for _, c := range copies {
		_, d, err := get(c[0])
		if err != nil {
			return err
		}
		i, _, err := get(c[1])
		if err != nil {
			return err
		}
		if err := big.replace(i, d); err != nil {
			return err
		}
	}
	for _, lang := range []string{"English", "Spanish", "French", "German", "Italian", "Largest", "Dutch"} {
		i, d, err := get("LANGUAGE_" + lang + "_Global_skate3ng.BIN")
		if err != nil {
			return err
		}
		if lang == "Dutch" {
			d = ar
		} else if d, _, err = rewriteTable(d, func(h uint32, s string) string {
			if h == langNameHash {
				return langNameAR
			}
			return s
		}); err != nil {
			return err
		}
		if err := big.replace(i, d); err != nil {
			return err
		}
	}
	tmp := archive + ".tmp"
	if err := os.WriteFile(tmp, big.data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, archive); err != nil {
		os.Remove(tmp)
		return fmt.Errorf(T("no pude reemplazar %s (¿el juego está abierto?): %v", "could not replace %s (is the game running?): %v"), textsArchive, err)
	}
	if err := os.WriteFile(archive+textsState, []byte("1"), 0o644); err != nil {
		return err
	}
	recordStamp(archive, "texts")
	c := loadConfig(root)
	c["texts"] = "1"
	if err := c.save(root); err != nil {
		return err
	}
	fmt.Printf(T("Listo: %d textos en rioplatense. Elegí \"Español ARG\" en la pantalla de idiomas del juego.\n",
		"Done: %d texts in Rioplatense. Pick \"Español ARG\" on the game's language screen.\n"), changed)
	return nil
}

func uninstallTexts(root, game string) error {
	archive, err := textsArchivePath(game)
	if err != nil {
		return err
	}
	backup := archive + textsBackup
	if exists(backup) {
		os.Remove(archive)
		if err := os.Rename(backup, archive); err != nil {
			return fmt.Errorf(T("no pude restaurar %s (¿el juego está abierto?): %v", "could not restore %s (is the game running?): %v"), textsArchive, err)
		}
	}
	os.Remove(archive + textsState)
	c := loadConfig(root)
	delete(c, "texts")
	if err := c.save(root); err != nil {
		return err
	}
	fmt.Println(T("Listo: idiomas originales restaurados (vuelve el neerlandés).", "Done: original languages restored (Dutch is back)."))
	return nil
}

func textsOn(game string) bool {
	archive, err := textsArchivePath(game)
	return err == nil && exists(archive+textsState) && exists(archive+textsBackup)
}
