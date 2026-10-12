package main

// Play time: how long the game was open each day (ini play_hist,
// "2026-10-12=5400;..." in seconds, last 60 days) and the last session for
// the summary the modder shows when you come back (ini last_session,
// "when|seconds|load ms").

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

const playDays = 60

func readPlayHist(c config) map[string]int {
	m := map[string]int{}
	for _, e := range strings.Split(c["play_hist"], ";") {
		if d, v, ok := strings.Cut(e, "="); ok {
			if n, err := strconv.Atoi(v); err == nil {
				m[d] = n
			}
		}
	}
	return m
}

// addPlayTime splits a session across the days it touched.
func addPlayTime(root string, start, end time.Time, load time.Duration) {
	if !end.After(start) {
		return
	}
	c := loadConfig(root)
	m := readPlayHist(c)
	for t := start; t.Before(end); {
		next := time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
		if next.After(end) {
			next = end
		}
		m[t.Format("2006-01-02")] += int(next.Sub(t) / time.Second)
		t = next
	}
	days := make([]string, 0, len(m))
	for d := range m {
		days = append(days, d)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(days)))
	if len(days) > playDays {
		days = days[:playDays]
	}
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, d+"="+strconv.Itoa(m[d]))
	}
	c["play_hist"] = strings.Join(parts, ";")
	c["last_session"] = end.Format("2006-01-02 15:04") + "|" + strconv.Itoa(int(end.Sub(start)/time.Second)) + "|" + strconv.Itoa(int(load/time.Millisecond))
	c.save(root)
}

// playStats is what the modder shows: today, this week (Monday on), the
// last 7 days one by one and the last session.
func playStats(root string, now time.Time) map[string]any {
	c := loadConfig(root)
	m := readPlayHist(c)
	today := now.Format("2006-01-02")
	wd := (int(now.Weekday()) + 6) % 7 // Monday = 0
	week, days := 0, []map[string]any{}
	for i := 0; i < 7; i++ {
		d := now.AddDate(0, 0, -i)
		k := d.Format("2006-01-02")
		if i <= wd {
			week += m[k]
		}
		days = append([]map[string]any{{"day": k, "wd": int(d.Weekday()), "secs": m[k]}}, days...)
	}
	total := 0
	for _, v := range m {
		total += v
	}
	out := map[string]any{"today": m[today], "week": week, "days": days, "total": total}
	if f := strings.Split(c["last_session"], "|"); len(f) == 3 {
		secs, _ := strconv.Atoi(f[1])
		load, _ := strconv.Atoi(f[2])
		out["last"] = map[string]any{"when": f[0], "secs": secs, "load": load}
	}
	return out
}
