package main

// Notice when skate3recomp itself has a new release: the installed version is
// read from the build stamp inside skate3.exe ("build: skate3-v2.0.2-..."),
// the latest one from GitHub. Each new version is announced until the user
// says "don't tell me about this one".

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const recompSlug = "mchughalex/skate3recomp"

func recompReleasesURL() string { return "https://github.com/" + recompSlug + "/releases/latest" }

var stampRe = regexp.MustCompile(`build: skate3-v([0-9]+\.[0-9]+(?:\.[0-9]+)*)`)

func recompExe(root string) string {
	for _, n := range []string{"skate3.exe", "skate3"} {
		if p := filepath.Join(root, n); exists(p) {
			return p
		}
	}
	return ""
}

// recompVersion finds the version in the game's exe. The result is cached in
// skate3-mods.ini against the exe's size and date, so the exe is scanned only
// after the recomp changes.
func recompVersion(root string) string {
	exe := recompExe(root)
	if exe == "" {
		return ""
	}
	st, err := os.Stat(exe)
	if err != nil {
		return ""
	}
	stamp := fmt.Sprintf("%d-%d", st.Size(), st.ModTime().Unix())
	c := loadConfig(root)
	if c["recomp_stamp"] == stamp {
		return c["recomp_version"]
	}
	v := scanStamp(exe)
	c["recomp_stamp"], c["recomp_version"] = stamp, v
	c.save(root)
	return v
}

// scanStamp reads the exe in chunks, keeping a small overlap so a stamp cut
// between two chunks is still found whole.
func scanStamp(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	const keep = 64
	buf := []byte{}
	chunk := make([]byte, 1<<20)
	for {
		n, err := io.ReadFull(f, chunk)
		buf = append(buf, chunk[:n]...)
		end := err != nil
		if loc := stampRe.FindSubmatchIndex(buf); loc != nil && (end || loc[1]+16 < len(buf)) {
			return string(buf[loc[2]:loc[3]])
		}
		if end {
			return ""
		}
		buf = append(buf[:0], buf[len(buf)-keep:]...)
	}
}

func latestRecomp(timeout time.Duration) (tag, url string, err error) {
	cl := http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+recompSlug+"/releases/latest", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Skate3-Mods/"+version)
	resp, err := cl.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("GitHub: %s", resp.Status)
	}
	var r struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", "", err
	}
	if !strings.HasPrefix(r.HTMLURL, "https://github.com/"+recompSlug+"/") {
		r.HTMLURL = recompReleasesURL()
	}
	return strings.TrimPrefix(r.TagName, "v"), r.HTMLURL, nil
}

// recompNews returns "installed|latest|url" when a newer recomp is out and
// the user did not skip it, or "" otherwise (also on any error).
func recompNews(root string) string {
	have := recompVersion(root)
	if have == "" {
		return ""
	}
	tag, url, err := latestRecomp(4 * time.Second)
	if err != nil || tag == "" || !versionLess(have, tag) || loadConfig(root)["recomp_skip"] == tag {
		return ""
	}
	return have + "|" + tag + "|" + url
}
