package main

// Load history: how long each direct load took and where the time went,
// read from the timestamps the game writes in its log.

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// bootPhases splits a direct load in four, from the moment Drop in was hit:
// open (the recomp starting until its first log line), menus (logos,
// language screen and intro until Press Start), save (your save and the
// world loading until you are skating) and spot (the trip to the starting
// spot, when there is one). Missing markers leave a phase at 0.
type bootPhases struct {
	Total, Open, Menus, Save, Spot time.Duration
}

var bootMarks = []struct {
	phase int // 1 = end of menus, 2 = end of save, 3 = end of spot
	text  string
}{
	{1, "enabled press-start auto-tap"},
	{2, "context=0x8001, value=0x1)"},
	{2, "gameplay reached"},
	{3, "gameplay input sequence complete"},
}

// logTime reads "[2026-10-12 20:01:02.345]" at the start of a log line.
func logTime(line string) (time.Time, bool) {
	if len(line) < 25 || line[0] != '[' {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05.000", line[1:24], time.Local)
	return t, err == nil
}

func readBootPhases(logPath string, start time.Time, total time.Duration) bootPhases {
	p := bootPhases{Total: total}
	f, err := os.Open(logPath)
	if err != nil {
		return p
	}
	defer f.Close()
	var first time.Time
	var ends [4]time.Time
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		t, ok := logTime(line)
		if !ok {
			continue
		}
		if first.IsZero() {
			first = t
		}
		for _, m := range bootMarks {
			if ends[m.phase].IsZero() && strings.Contains(line, m.text) {
				ends[m.phase] = t
			}
		}
	}
	if first.IsZero() {
		return p
	}
	if d := first.Sub(start); d > 0 {
		p.Open = d
	}
	if !ends[1].IsZero() {
		p.Menus = ends[1].Sub(first)
	}
	if !ends[2].IsZero() {
		from := ends[1]
		if from.IsZero() {
			from = first
		}
		p.Save = ends[2].Sub(from)
	}
	if !ends[3].IsZero() && !ends[2].IsZero() {
		p.Spot = ends[3].Sub(ends[2])
	}
	return p
}

const bootHistMax = 20

// addBootHistory keeps the last loads in the ini, newest first:
// "date|total|open|menus|save|spot;..." in milliseconds.
func addBootHistory(root string, when time.Time, p bootPhases) {
	ms := func(d time.Duration) string { return strconv.Itoa(int(d / time.Millisecond)) }
	e := strings.Join([]string{when.Format("2006-01-02 15:04"), ms(p.Total), ms(p.Open), ms(p.Menus), ms(p.Save), ms(p.Spot)}, "|")
	c := loadConfig(root)
	h := []string{e}
	for _, old := range strings.Split(c["boot_hist"], ";") {
		if old != "" && len(h) < bootHistMax {
			h = append(h, old)
		}
	}
	c["boot_hist"] = strings.Join(h, ";")
	c.save(root)
}

// bootHistory is the history for the modder, newest first.
func bootHistory(root string) []map[string]any {
	out := []map[string]any{}
	for _, e := range strings.Split(loadConfig(root)["boot_hist"], ";") {
		f := strings.Split(e, "|")
		if len(f) != 6 {
			continue
		}
		n := func(s string) int { v, _ := strconv.Atoi(s); return v }
		out = append(out, map[string]any{"when": f[0], "total": n(f[1]), "open": n(f[2]), "menus": n(f[3]), "save": n(f[4]), "spot": n(f[5])})
	}
	return out
}
