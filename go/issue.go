package main

// GitHub issues: once the pack's repository is public, a problem report or
// a translation fix opens GitHub's "new issue" page already filled in.

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	repoOnce   sync.Once
	repoPublic bool
)

// repoIsPublic asks GitHub once per run (in the background from the start).
func repoIsPublic() bool {
	repoOnce.Do(func() {
		cl := http.Client{Timeout: 6 * time.Second}
		req, err := http.NewRequest("GET", "https://api.github.com/repos/"+repoSlug, nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", "Skate3-Mods/"+version)
		if resp, err := cl.Do(req); err == nil {
			resp.Body.Close()
			repoPublic = resp.StatusCode == 200
		}
	})
	return repoPublic
}

// issueURL builds the prefilled "new issue" link; GitHub takes URLs of about
// 8 KB, so the body is cut to fit.
func issueURL(title, body string) string {
	const max = 6000
	if len(body) > max {
		body = body[:max] + "\n..."
	}
	return "https://github.com/" + repoSlug + "/issues/new?title=" + url.QueryEscape(title) + "&body=" + url.QueryEscape(body)
}

// openIssue fills the report and opens it on GitHub.
func openIssue(root, kind, what string) (string, error) {
	title := strings.TrimSpace(what)
	if i := strings.IndexByte(title, '\n'); i > 0 {
		title = title[:i]
	}
	if len(title) > 70 {
		title = title[:70] + "..."
	}
	body := "**" + T("Qué pasó", "What happened") + "**\n\n" + strings.TrimSpace(what) + "\n\n```\n" + diagSummary(root) + "\n```"
	if kind == "translation" {
		title = "[" + lang + "] " + title
		body = "**Translation fix (" + lang + ")**\n\n" + strings.TrimSpace(what) + "\n\nSkate 3 Mods v" + version
	}
	return T("Se abrió GitHub con el reporte listo para enviar", "GitHub opened with the report ready to send"), openPath(issueURL(title, body))
}

// diagSummary is the short system part of the report.
func diagSummary(root string) string {
	c := loadConfig(root)
	return strings.Join([]string{
		"Skate 3 Mods v" + version + " / recomp " + c["recomp_version"],
		systemInfo(),
		T("Iconos: ", "Icons: ") + iconsKey(root) + " / preset: " + c["preset"] + " / direct: " + onOff(directOn(root)) + " / lang: " + lang,
	}, "\n")
}
