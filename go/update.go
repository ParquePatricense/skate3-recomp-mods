package main

// Updates: asks GitHub for the latest release of the pack, says so when it
// is newer than this build, and can install it in one go. Silent on errors
// while checking.

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const repoSlug = "ParquePatricense/skate3-recomp-mods"

func releasesURL() string { return "https://github.com/" + repoSlug + "/releases/latest" }

type release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func latestRelease(timeout time.Duration) (*release, error) {
	cl := http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+repoSlug+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Skate3-Mods/"+version)
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub: %s", resp.Status)
	}
	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// selfUpdate downloads the latest Skate3-Mods.zip, swaps the running exe for
// the new one (Windows allows renaming a running exe) and starts it.
func selfUpdate() error {
	r, err := latestRelease(10 * time.Second)
	if err != nil {
		return err
	}
	if !versionLess(version, r.TagName) {
		return errors.New(T("ya tenés la última versión", "you already have the latest version"))
	}
	url := ""
	for _, a := range r.Assets {
		if strings.EqualFold(a.Name, "Skate3-Mods.zip") {
			url = a.URL
		}
	}
	if url == "" {
		return errors.New(T("la versión nueva no trae Skate3-Mods.zip", "the new release has no Skate3-Mods.zip"))
	}
	cl := http.Client{Timeout: 2 * time.Minute}
	resp, err := cl.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	var newExe []byte
	for _, f := range zr.File {
		if strings.EqualFold(f.Name, modsExe) {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			newExe, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}
		}
	}
	if len(newExe) < 1<<20 {
		return errors.New(T("el zip descargado no trae el programa", "the downloaded zip has no program in it"))
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.WriteFile(self+".new", newExe, 0o755); err != nil {
		return err
	}
	os.Remove(self + ".old")
	if err := os.Rename(self, self+".old"); err != nil {
		return err
	}
	if err := os.Rename(self+".new", self); err != nil {
		os.Rename(self+".old", self)
		return err
	}
	return exec.Command(self).Start()
}

// cleanOldExe removes what selfUpdate left behind.
func cleanOldExe() {
	if self, err := os.Executable(); err == nil {
		os.Remove(self + ".old")
	}
}

// newerVersion returns the latest release tag when it is newer than version.
func newerVersion() string {
	r, err := latestRelease(3 * time.Second)
	if err != nil {
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
