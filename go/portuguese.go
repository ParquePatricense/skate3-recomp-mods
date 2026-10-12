package main

// Português (Brasil): the game's texts translated from the Spanish table,
// one line per string ("hash<TAB>text"), shipped compressed. It goes into the
// English slot like the other boot languages, so it needs no new language
// screen entry; that slot's "English" shows as "Português".

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"strconv"
	"strings"
	"sync"
)

//go:embed ptbr.tsv.gz
var ptbrGz []byte

const langNameEnglish = 0x4d866eb8 // ID_LANGUAGE_ENGLISH

var (
	ptOnce sync.Once
	ptMap  map[uint32]string
)

func ptTexts() map[uint32]string {
	ptOnce.Do(func() {
		ptMap = map[uint32]string{}
		r, err := gzip.NewReader(bytes.NewReader(ptbrGz))
		if err != nil {
			return
		}
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 64<<10), 1<<20)
		for sc.Scan() {
			h, t, ok := strings.Cut(sc.Text(), "\t")
			if v, err := strconv.ParseUint(h, 16, 32); ok && err == nil {
				ptMap[uint32(v)] = t
			}
		}
	})
	return ptMap
}

// portuguese rewrites the Spanish table into Portuguese; anything missing
// stays in Spanish.
func portuguese(es []byte) ([]byte, error) {
	pt := ptTexts()
	d, _, err := rewriteTable(es, func(h uint32, s string) string {
		if h == langNameEnglish {
			return "Português"
		}
		if t, ok := pt[h]; ok {
			return t
		}
		return s
	})
	return d, err
}
