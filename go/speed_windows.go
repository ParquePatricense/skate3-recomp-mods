//go:build windows

package main

// Faster loads, outside the game: Windows Defender skipping the game folder,
// the game files read into memory when Windows starts, and moving the game
// to an SSD. Also the stick drift check (XInput).

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// ---------------------------------------------------------------- Defender

// setDefender asks Windows (with its admin prompt) to leave the game folder
// and skate3.exe out of Defender's scans, or to put them back.
func setDefender(root string, on bool) (string, error) {
	verb := "Add-MpPreference"
	if !on {
		verb = "Remove-MpPreference"
	}
	inner := verb + " -ExclusionPath " + psQuote(root) + " -ExclusionProcess " + psQuote(filepath.Join(root, gameExe))
	script := "$p=Start-Process powershell -Verb RunAs -Wait -PassThru -WindowStyle Hidden -ArgumentList " +
		psQuote(`-NoProfile -NonInteractive -WindowStyle Hidden -Command "& {`+inner+`; if(-not $?){exit 1}}"`) + ";exit $p.ExitCode"
	if err := powershell(script); err != nil {
		return "", errors.New(T("Windows no aceptó el cambio (¿dijiste que no al permiso de administrador?)", "Windows didn't take the change (did you say no to the admin prompt?)"))
	}
	c := loadConfig(root)
	c["defender"] = map[bool]string{true: "1", false: "0"}[on]
	if err := c.save(root); err != nil {
		return "", err
	}
	if on {
		return T("Listo: Defender ya no revisa el juego al cargar", "Done: Defender no longer scans the game while it loads"), nil
	}
	return T("Defender vuelve a revisar el juego", "Defender scans the game again"), nil
}

// ---------------------------------------------------------------- preload

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const runName = "Skate3ModsPreload"

func preloadOn() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runName)
	return err == nil
}

// setPreload: Windows runs "Skate3-Mods.exe preload" at sign-in, which
// reads the game files into memory in the background and exits.
func setPreload(root string, on bool) (string, error) {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	if !on {
		k.DeleteValue(runName)
		return T("Precarga al prender la PC: no", "Preload at startup: off"), nil
	}
	exe := filepath.Join(root, modsExe)
	if !exists(exe) {
		if exe, err = os.Executable(); err != nil {
			return "", err
		}
	}
	if err := k.SetStringValue(runName, `"`+exe+`" preload --game "`+root+`"`); err != nil {
		return "", err
	}
	return T("Precarga al prender la PC: sí", "Preload at startup: on"), nil
}

// preload is the sign-in run: it waits a bit so it doesn't slow the
// desktop down, then warms the game files and exits.
func preload(root string) {
	time.Sleep(45 * time.Second)
	warmRun(root)
}

// ---------------------------------------------------------------- SSD

// driveIsSSD asks the disk under a path whether it has a seek penalty
// (spinning disks do). ok is false when Windows can't tell.
func driveIsSSD(path string) (ssd, ok bool) {
	vol := filepath.VolumeName(path)
	if vol == "" {
		return false, false
	}
	h, err := windows.CreateFile(utf16(`\\.\`+vol), 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return false, false
	}
	defer windows.CloseHandle(h)
	const ioctlStorageQueryProperty = 0x2D1400
	query := [3]uint32{7, 0, 0} // StorageDeviceSeekPenaltyProperty, PropertyStandardQuery
	var out [3]uint32           // Version, Size, IncursSeekPenalty (BOOLEAN)
	var n uint32
	if windows.DeviceIoControl(h, ioctlStorageQueryProperty, (*byte)(unsafe.Pointer(&query[0])), 12, (*byte)(unsafe.Pointer(&out[0])), 12, &n, nil) != nil || n < 9 {
		return false, false
	}
	return out[2]&0xFF == 0, true
}

type driveInfo struct {
	Letter string `json:"letter"`
	Free   uint64 `json:"free"`
	SSD    bool   `json:"ssd"`
}

// otherSSDs lists the fixed SSD drives other than the game's.
func otherSSDs(root string) []driveInfo {
	var out []driveInfo
	cur := strings.ToUpper(filepath.VolumeName(root))
	mask, _ := windows.GetLogicalDrives()
	for i := 0; i < 26; i++ {
		if mask&(1<<i) == 0 {
			continue
		}
		l := string(rune('A'+i)) + ":"
		if l == cur || windows.GetDriveType(utf16(l+`\`)) != windows.DRIVE_FIXED {
			continue
		}
		if ssd, ok := driveIsSSD(l + `\`); ok && ssd {
			free, _ := freeSpace(l + `\`)
			out = append(out, driveInfo{Letter: l, Free: free, SSD: true})
		}
	}
	return out
}

func dirSize(root string) int64 {
	var n int64
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if st, e := d.Info(); e == nil {
				n += st.Size()
			}
		}
		return nil
	})
	return n
}

// diskState is what the modder shows about the game's disk.
func diskState(root string) map[string]any {
	ssd, ok := driveIsSSD(root)
	st := map[string]any{"known": ok, "ssd": ssd, "drive": filepath.VolumeName(root)}
	if ok && !ssd {
		st["targets"] = otherSSDs(root)
	}
	moveMu.Lock()
	if moveJob != nil {
		st["move"] = map[string]any{"done": moveJob.done, "total": moveJob.total, "err": moveJob.err, "to": moveJob.to, "finished": moveJob.finished}
	}
	moveMu.Unlock()
	if old := loadConfig(root)["moved_from"]; old != "" && exists(old) {
		st["movedFrom"] = old
	}
	return st
}

type moveState struct {
	done, total int64
	to, err     string
	finished    bool
}

var (
	moveMu  sync.Mutex
	moveJob *moveState
)

// startMove copies the whole game folder to <drive>\Skate3Recomp in the
// background. The old folder stays until you delete it from the modder.
func startMove(root, drive string) (string, error) {
	moveMu.Lock()
	defer moveMu.Unlock()
	if moveJob != nil && !moveJob.finished {
		return "", errors.New(T("ya estoy moviendo el juego", "already moving the game"))
	}
	if _, running := findGamePID(); running {
		return "", errors.New(T("cerrá el juego primero", "close the game first"))
	}
	ok := false
	for _, d := range otherSSDs(root) {
		ok = ok || strings.EqualFold(d.Letter, drive)
	}
	if !ok {
		return "", errors.New("?")
	}
	to := filepath.Join(drive+`\`, "Skate3Recomp")
	for i := 2; exists(to); i++ {
		to = filepath.Join(drive+`\`, fmt.Sprintf("Skate3Recomp %d", i))
	}
	total := dirSize(root)
	if free, _ := freeSpace(drive + `\`); free < uint64(total)+(1<<30) {
		return "", fmt.Errorf(T("no entra: hacen falta %.1f GB libres en %s", "it doesn't fit: %.1f GB free needed on %s"), float64(total)/(1<<30)+1, drive)
	}
	job := &moveState{total: total, to: to}
	moveJob = job
	go func() {
		err := copyTree(root, to, func(n int64) {
			moveMu.Lock()
			job.done += n
			moveMu.Unlock()
		})
		if err == nil {
			err = finishMove(root, to)
		}
		moveMu.Lock()
		if err != nil {
			job.err = err.Error()
			os.RemoveAll(to)
		}
		job.finished = true
		moveMu.Unlock()
	}()
	return T("Moviendo el juego a ", "Moving the game to ") + to, nil
}

func copyTree(from, to string, progress func(int64)) error {
	buf := make([]byte, 4<<20)
	return filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, p)
		dst := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		for {
			n, rerr := in.Read(buf)
			if n > 0 {
				if _, werr := out.Write(buf[:n]); werr != nil {
					out.Close()
					return werr
				}
				progress(int64(n))
			}
			if rerr == io.EOF {
				break
			}
			if rerr != nil {
				out.Close()
				return rerr
			}
		}
		if err := out.Close(); err != nil {
			return err
		}
		if st, e := d.Info(); e == nil {
			os.Chtimes(dst, st.ModTime(), st.ModTime())
		}
		return nil
	})
}

// finishMove points everything at the new folder: the remembered game
// folder, the shortcuts and the Steam entry.
func finishMove(old, to string) error {
	if dirSize(to) != dirSize(old) {
		return errors.New(T("la copia no salió completa", "the copy came out incomplete"))
	}
	c := loadConfig(to)
	c["moved_from"] = old
	if err := c.save(to); err != nil {
		return err
	}
	rememberRoot(to)
	if c.on("shortcut") {
		createShortcut(to)
	}
	if c.on("steam") {
		setSteam(to, true)
	}
	if preloadOn() {
		setPreload(to, true)
	}
	return nil
}

// deleteOldGame removes the folder the game was moved from.
func deleteOldGame(root string) (string, error) {
	c := loadConfig(root)
	old := c["moved_from"]
	if old == "" || !exists(filepath.Join(old, gameExe)) || strings.EqualFold(filepath.Clean(old), filepath.Clean(root)) {
		return "", errors.New("?")
	}
	if err := os.RemoveAll(old); err != nil {
		return "", err
	}
	delete(c, "moved_from")
	return T("Carpeta vieja borrada", "Old folder deleted"), c.save(root)
}

// ---------------------------------------------------------------- drift

var pXInputGetState = syscall.NewLazyDLL("xinput1_4.dll").NewProc("XInputGetState")

type xinputState struct {
	Packet  uint32
	Buttons uint16
	LT, RT  uint8
	LX, LY  int16
	RX, RY  int16
}

// stickDrift samples the pads for about a second with nobody touching them
// (no buttons, sticks still): a stick that rests off-center past the
// game's dead zone drifts. It returns one line per drifting stick.
func stickDrift() []map[string]any {
	if pXInputGetState.Find() != nil {
		return nil
	}
	const samples = 20
	const dead = 7849 // XInput's own left-stick dead zone
	var sum [4][4]int64
	var seen [4]int
	for s := 0; s < samples; s++ {
		for pad := 0; pad < 4; pad++ {
			var st xinputState
			if r, _, _ := pXInputGetState.Call(uintptr(pad), uintptr(unsafe.Pointer(&st))); r != 0 {
				continue
			}
			if st.Buttons != 0 || st.LT > 30 || st.RT > 30 {
				continue // someone's holding it
			}
			seen[pad]++
			for i, v := range []int16{st.LX, st.LY, st.RX, st.RY} {
				sum[pad][i] += int64(v)
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	var out []map[string]any
	for pad := 0; pad < 4; pad++ {
		if seen[pad] < samples/2 {
			continue
		}
		for i, name := range []string{"left", "right"} {
			x, y := sum[pad][i*2]/int64(seen[pad]), sum[pad][i*2+1]/int64(seen[pad])
			mag := x*x + y*y
			if mag > dead*dead {
				pct := int(math.Sqrt(float64(mag)) * 100 / 32767)
				out = append(out, map[string]any{"pad": pad + 1, "stick": name, "pct": pct})
			}
		}
	}
	return out
}

func padsConnected() int {
	if pXInputGetState.Find() != nil {
		return 0
	}
	n := 0
	for pad := 0; pad < 4; pad++ {
		var st xinputState
		if r, _, _ := pXInputGetState.Call(uintptr(pad), uintptr(unsafe.Pointer(&st))); r == 0 {
			n++
		}
	}
	return n
}
