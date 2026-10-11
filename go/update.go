package main

// Update notice: asks GitHub for the latest release of the pack and says so
// in the menu when it is newer than this build. Silent on any error.

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const repoSlug = "ParquePatricense/skate3-recomp-mods"

func releasesURL() string { return "https://github.com/" + repoSlug + "/releases/latest" }

// newerVersion returns the latest release tag when it is newer than version.
func newerVersion() string {
	cl := http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+repoSlug+"/releases/latest", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Skate3-Mods/"+version)
	resp, err := cl.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}
	var r struct {
		TagName string `json:"tag_name"`
	}
	if json.NewDecoder(resp.Body).Decode(&r) != nil {
		return ""
	}
	if versionLess(version, r.TagName) {
		return strings.TrimPrefix(r.TagName, "v")
	}
	return ""
}

func versionLess(a, b string) bool {
	pa := strings.Split(strings.TrimPrefix(a, "v"), ".")
	pb := strings.Split(strings.TrimPrefix(b, "v"), ".")
	for i := 0; i < 3; i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			return x < y
		}
	}
	return false
}
