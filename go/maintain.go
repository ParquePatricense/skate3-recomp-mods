// Maintenance: put the mods back after a recomp update, repair broken
// files, uninstall everything, diagnostics and handy folders.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

// A patched archive and how to rebuild it.
type modFile struct {
	key     string // ini key prefix for its stamp
	name    string // archive file name
	entry   string // entry the mod changes (used to check the file is healthy)
	backup  string // backup suffix
	state   string // "installed" marker suffix
	install func(root string) error
	patched func(root string) bool // is the archive exactly what the mod writes?
}

func modFiles() []modFile {
	return []modFile{
		{"icons", archiveName, entryName, backupSuffix, stateSuffix, func(root string) error {
			st, _ := os.ReadFile(filepath.Join(bigDir(root), archiveName+stateSuffix))
			return installIcons(strings.TrimSpace(string(st)), root)
		}, func(root string) bool {
			st, _ := os.ReadFile(filepath.Join(bigDir(root), archiveName+stateSuffix))
			p := filepath.Join(bigDir(root), archiveName)
			orig, err1 := readEntry(p+backupSuffix, entryName)
			cur, err2 := readEntry(p, entryName)
			if err1 != nil || err2 != nil {
				return false
			}
			want, err := patchRx2(orig, strings.TrimSpace(string(st)))
			return err == nil && bytes.Equal(cur, want)
		}},
		{"texts", textsArchive, textsEntry, textsBackup, textsState, func(root string) error {
			_, err := applyTexts(root, root)
			return err
		}, func(root string) bool {
			return textsPatched(root)
		}},
	}
}

func bigDir(root string) string { return filepath.Join(root, "game", "data", "big") }

func fileStamp(p string) string {
	st, err := os.Stat(p)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d:%d", st.Size(), st.ModTime().UnixNano())
}

// recordStamp remembers how a file looked right after the mod wrote it.
func recordStamp(archive, key string) {
	root, err := findRoot(filepath.Dir(filepath.Dir(filepath.Dir(archive))))
	if err != nil {
		return
	}
	c := loadConfig(root)
	c[key+"_stamp"] = fileStamp(archive)
	c.save(root)
}

func readEntry(p, entry string) ([]byte, error) {
	d, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	b, err := parseBig(d)
	if err != nil {
		return nil, err
	}
	i, err := b.find(entry)
	if err != nil {
		return nil, err
	}
	return b.read(i)
}

func archiveHealthy(p, entry string) bool {
	d, err := os.ReadFile(p)
	if err != nil {
		return false
	}
	b, err := parseBig(d)
	if err != nil {
		return false
	}
	i, err := b.find(entry)
	if err != nil {
		return false
	}
	_, err = b.read(i)
	return err == nil
}

// checkMods runs before playing and when the modder opens. If a module is
// on but its file changed behind our back, either the recomp (or a
// reinstall) put a fresh original there, which becomes the new backup, or
// the file is broken and comes back from the backup. Then the mod is
// applied again. force also checks files that look untouched (Repair).
func checkMods(root string) []string {
	var out []string
	for _, m := range modFiles() {
		p := filepath.Join(bigDir(root), m.name)
		backup, state := p+m.backup, p+m.state
		if !exists(state) {
			continue
		}
		if !exists(backup) {
			os.Remove(state)
			continue
		}
		c := loadConfig(root)
		rec := c[m.key+"_stamp"]
		if rec == "" {
			c[m.key+"_stamp"] = fileStamp(p)
			c.save(root)
			continue
		}
		if fileStamp(p) == rec && archiveHealthy(p, m.entry) {
			continue
		}
		if archiveHealthy(p, m.entry) && m.patched(root) {
			// still our file (copied or touched): just remember it again
			c[m.key+"_stamp"] = fileStamp(p)
			c.save(root)
			continue
		}
		if archiveHealthy(p, m.entry) {
			if err := copyFile(p, backup); err != nil {
				out = append(out, m.name+": "+err.Error())
				continue
			}
			out = append(out, fmt.Sprintf(T("%s cambió (¿actualizaste el recomp?): lo tomé como nuevo original.",
				"%s changed (did you update the recomp?): it is the new original now."), m.name))
		} else if err := copyFile(backup, p); err != nil {
			out = append(out, m.name+": "+err.Error())
			continue
		} else {
			out = append(out, fmt.Sprintf(T("%s estaba roto: lo restauré desde la copia.", "%s was broken: restored it from the backup."), m.name))
		}
		if err := m.install(root); err != nil {
			out = append(out, m.name+": "+err.Error())
		} else {
			out = append(out, fmt.Sprintf(T("%s: mod aplicado de nuevo.", "%s: mod applied again."), m.name))
		}
	}
	return out
}

// repairMods checks every archive the pack touches, even if untouched.
func repairMods(root string) []string {
	out := checkMods(root)
	for _, m := range modFiles() {
		p := filepath.Join(bigDir(root), m.name)
		backup := p + m.backup
		if archiveHealthy(p, m.entry) {
			continue
		}
		if !exists(backup) || !archiveHealthy(backup, m.entry) {
			out = append(out, fmt.Sprintf(T("%s está roto y no hay copia sana: reinstalá el juego desde el ISO (Select ISO).",
				"%s is broken and there is no healthy backup: reinstall the game from the ISO (Select ISO)."), m.name))
			continue
		}
		if err := copyFile(backup, p); err != nil {
			out = append(out, m.name+": "+err.Error())
			continue
		}
		msg := fmt.Sprintf(T("%s estaba roto: lo restauré.", "%s was broken: restored it."), m.name)
		if exists(p + m.state) {
			if err := m.install(root); err != nil {
				msg += " " + err.Error()
			}
		}
		out = append(out, msg)
	}
	if len(out) == 0 {
		out = append(out, T("Todo en orden: los archivos del juego están sanos.", "All good: the game files are healthy."))
	}
	return out
}

// uninstallAll leaves the game exactly as it was before the pack.
func uninstallAll(root string) []string {
	var out []string
	add := func(what string, err error) {
		if err != nil {
			out = append(out, what+": "+err.Error())
		}
	}
	add(T("iconos", "icons"), uninstallIcons(root))
	c := loadConfig(root)
	if c.on("steam") {
		_, err := setSteam(root, false)
		add("Steam", err)
		c = loadConfig(root)
	}
	delete(c, "texts")
	delete(c, "boot_lang")
	c.save(root)
	_, err := applyTexts(root, root)
	add(T("idioma", "language"), err)
	add("skate3.toml", restoreToml(root))
	add(T("acceso directo", "shortcut"), removeShortcut())
	dndOff(root)
	os.Remove(filepath.Join(root, modsIni))
	if f := rememberFile(); f != "" {
		os.Remove(f)
	}
	out = append(out, T("Listo: el juego quedó como estaba antes de los mods. Podés borrar Skate3-Mods.exe y el .bat.",
		"Done: the game is back to how it was before the mods. You can delete Skate3-Mods.exe and the .bat."))
	return out
}

// savesDir is where the recomp keeps profiles and saves.
func savesDir(root string) string {
	if b, err := os.ReadFile(filepath.Join(root, tomlName)); err == nil {
		if m := regexp.MustCompile(`(?m)^\s*user_data_root\s*=\s*"([^"]+)"`).FindSubmatch(b); m != nil {
			return filepath.Join(string(m[1]), "saves")
		}
	}
	return filepath.Join(documentsDir(), "skate3", "saves")
}

func folderPath(root, which string) string {
	switch which {
	case "saves":
		return savesDir(root)
	case "data":
		return bigDir(root)
	case "userdata":
		return filepath.Dir(savesDir(root))
	}
	return root
}

func openFolder(root, which string) error {
	p := folderPath(root, which)
	if !exists(p) {
		return fmt.Errorf(T("todavía no existe %s (se crea al jugar)", "%s doesn't exist yet (it is created when you play)"), p)
	}
	return openPath(p)
}

// diagnose writes a plain-text report to the desktop and opens it.
func diagnose(root string) (string, error) {
	var b strings.Builder
	w := func(f string, a ...any) { fmt.Fprintf(&b, f+"\n", a...) }
	w("Skate 3 Recomp Mods v%s - %s", version, time.Now().Format("2006-01-02 15:04"))
	w("Sistema: %s/%s %s", runtime.GOOS, runtime.GOARCH, systemInfo())
	exe, _ := os.Executable()
	w("Programa: %s", exe)
	w("Carpeta del juego: %s", root)
	w("")
	w("== Archivos")
	for _, f := range []string{"skate3.exe", tomlName, modsIni,
		filepath.Join("game", "data", "big", archiveName), filepath.Join("game", "data", "big", archiveName+backupSuffix),
		filepath.Join("game", "data", "big", textsArchive), filepath.Join("game", "data", "big", textsArchive+textsBackup)} {
		if st, err := os.Stat(filepath.Join(root, f)); err == nil {
			w("  OK  %-45s %12d bytes  %s", f, st.Size(), st.ModTime().Format("2006-01-02 15:04"))
		} else {
			w("  --  %s (no está)", f)
		}
	}
	w("")
	w("== Chequeo de archivos del juego")
	for _, m := range modFiles() {
		p := filepath.Join(bigDir(root), m.name)
		w("  %-16s sano: %v   mod activo: %v", m.name, archiveHealthy(p, m.entry), exists(p+m.state))
	}
	icons, _ := iconsStatus(root)
	w("  Iconos: %s   Español ARG: %v", icons, textsOn(root))
	w("")
	w("== skate-mods.ini")
	c := loadConfig(root)
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		w("  %s = %s", k, c[k])
	}
	w("")
	w("== skate3.toml (valores que tocan los mods)")
	if t, err := os.ReadFile(filepath.Join(root, tomlName)); err == nil {
		re := regexp.MustCompile(`^\s*(skate3_|anisotropic|fullscreen|window_|monitor|resolution|user_language|vsync|mnk_|keybind_|present_)`)
		for _, ln := range strings.Split(string(t), "\n") {
			if re.MatchString(ln) {
				w("  %s", strings.TrimSpace(ln))
			}
		}
	} else {
		w("  (no hay skate3.toml todavía)")
	}
	w("")
	sd := savesDir(root)
	n := 0
	if es, err := os.ReadDir(sd); err == nil {
		n = len(es)
	}
	w("== Partidas: %s (%d elementos)", sd, n)
	p := filepath.Join(desktopDir(), "Skate3-Mods-diagnostico.txt")
	if err := os.WriteFile(p, []byte(strings.ReplaceAll(b.String(), "\n", "\r\n")), 0o644); err != nil {
		return "", err
	}
	openPath(p)
	return p, nil
}

// problemReport is the diagnostics plus what the user says happened, saved on the
// desktop and copied so it can be pasted straight into a message.
// newestLog is the latest *.log in dir written after since.
func newestLog(dir string, since time.Time) string {
	ents, _ := os.ReadDir(dir)
	best, bt := "", since
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".log") {
			continue
		}
		if i, err := e.Info(); err == nil && i.ModTime().After(bt) {
			best, bt = filepath.Join(dir, e.Name()), i.ModTime()
		}
	}
	return best
}

// logTail is the end of the newest game log, for the problem report.
func logTail(root string, lines int) string {
	p := newestLog(filepath.Join(root, "logs"), time.Time{})
	if p == "" {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	l := strings.Split(strings.TrimRight(string(b), "\r\n"), "\n")
	if len(l) > lines {
		l = l[len(l)-lines:]
	}
	return "\n== " + T("Registro del juego", "Game log") + " (" + filepath.Base(p) + ")\n" + strings.Join(l, "\n") + "\n"
}

func problemReport(root, what string) (string, error) {
	p, err := diagnose(root)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	text := "== " + T("Qué pasó", "What happened") + "\n" + strings.TrimSpace(what) + "\n\n" + string(b) + logTail(root, 80)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		return "", err
	}
	if copyToClipboard(p) == nil {
		return T("Reporte copiado: pegalo en el chat. También quedó en ", "Report copied: paste it in the chat. It is also in ") + p, nil
	}
	return T("Reporte guardado en ", "Report saved to ") + p, nil
}
