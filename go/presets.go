package main

// Graphics presets: write known skate3recomp cvars into skate3.toml. The
// user's own skate3.toml is backed up once and can be restored at any time.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	tomlName      = "skate3.toml"
	tomlBackup    = "skate3.toml.mods-original"
	tomlNoOrigTag = "# (no habia skate3.toml)\n"
)

type preset struct {
	key, name string
	values    [][2]string
}

var presets = []preset{
	{"max", "Máxima calidad|Maximum quality", [][2]string{
		{"skate3_native_render_scene", "true"},
		{"skate3_native_render_scene_msaa", "8"},
		{"skate3_native_render_scene_shadows", "true"},
		{"skate3_native_render_scene_shadow_static_casters", "true"},
		{"skate3_native_render_scene_shadow_pcss", "true"},
		{"skate3_native_render_scene_ssao", "true"},
		{"skate3_native_render_scene_bloom", "true"},
		{"skate3_native_render_scene_shafts", "true"},
		{"skate3_native_render_scene_haze", "true"},
		{"skate3_draw_distance_scale", "4.0"},
		{"skate3_lod_distance_scale", "4.0"},
		{"anisotropic_override", "16"},
		{"skate3_native_render_scene_shadow_static_size", "4096"},
		{"skate3_native_render_scene_shafts_res", "2"},
	}},
	{"balanced", "Equilibrado (valores por defecto del recomp)|Balanced (recomp defaults)", [][2]string{
		{"skate3_native_render_scene", "true"},
		{"skate3_native_render_scene_msaa", "4"},
		{"skate3_native_render_scene_shadows", "true"},
		{"skate3_native_render_scene_shadow_static_casters", "true"},
		{"skate3_native_render_scene_shadow_pcss", "true"},
		{"skate3_native_render_scene_ssao", "true"},
		{"skate3_native_render_scene_bloom", "true"},
		{"skate3_native_render_scene_shafts", "true"},
		{"skate3_native_render_scene_haze", "true"},
		{"skate3_draw_distance_scale", "2.0"},
		{"skate3_lod_distance_scale", "2.0"},
		{"anisotropic_override", "-1"},
		{"skate3_native_render_scene_shadow_static_size", "4096"},
		{"skate3_native_render_scene_shafts_res", "4"},
	}},
	{"perf", "Rendimiento (PC de pocos recursos)|Performance (low-end PCs)", [][2]string{
		{"skate3_native_render_scene", "true"},
		{"skate3_native_render_scene_msaa", "1"},
		{"skate3_native_render_scene_shadows", "true"},
		{"skate3_native_render_scene_shadow_static_casters", "false"},
		{"skate3_native_render_scene_shadow_pcss", "false"},
		{"skate3_native_render_scene_ssao", "false"},
		{"skate3_native_render_scene_bloom", "false"},
		{"skate3_native_render_scene_shafts", "false"},
		{"skate3_native_render_scene_haze", "false"},
		{"skate3_draw_distance_scale", "1.0"},
		{"skate3_lod_distance_scale", "1.0"},
		{"anisotropic_override", "-1"},
		{"skate3_native_render_scene_shadow_static_size", "2048"},
		{"skate3_native_render_scene_shafts_res", "8"},
	}},
	{"ultra", "Ultra liviano (PC muy justa)|Ultra light (very weak PCs)", [][2]string{
		{"skate3_native_render_scene", "true"},
		{"skate3_native_render_scene_msaa", "1"},
		{"skate3_native_render_scene_shadows", "false"},
		{"skate3_native_render_scene_shadow_static_casters", "false"},
		{"skate3_native_render_scene_shadow_pcss", "false"},
		{"skate3_native_render_scene_ssao", "false"},
		{"skate3_native_render_scene_bloom", "false"},
		{"skate3_native_render_scene_shafts", "false"},
		{"skate3_native_render_scene_haze", "false"},
		{"skate3_draw_distance_scale", "0.75"},
		{"skate3_lod_distance_scale", "0.75"},
		{"anisotropic_override", "-1"},
		{"skate3_native_render_scene_shadow_static_size", "1024"},
		{"skate3_native_render_scene_shafts_res", "8"},
	}},
}

// label returns the preset name in the UI language ("es|en").
func (p preset) label() string {
	es, en, _ := strings.Cut(p.name, "|")
	return T(es, en)
}

func findPreset(key string) *preset {
	for i := range presets {
		if presets[i].key == key {
			return &presets[i]
		}
	}
	return nil
}

// setTomlValues rewrites "key = ..." lines (or appends them) keeping the rest.
func setTomlValues(root string, values [][2]string) error {
	path := filepath.Join(root, tomlName)
	backup := filepath.Join(root, tomlBackup)
	old, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !exists(backup) {
		orig := old
		if err != nil {
			orig = []byte(tomlNoOrigTag)
		}
		if err := os.WriteFile(backup, orig, 0o644); err != nil {
			return err
		}
	}
	text := string(old)
	for _, kv := range values {
		re := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(kv[0]) + `[ \t]*=.*$`)
		line := kv[0] + " = " + kv[1]
		if re.MatchString(text) {
			text = re.ReplaceAllLiteralString(text, line)
		} else {
			if text != "" && !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			text += line + "\n"
		}
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

func applyPreset(root, key string) error {
	p := findPreset(key)
	if p == nil {
		return fmt.Errorf(T("preset desconocido %q (usá max, balanced, perf o ultra)", "unknown preset %q (use max, balanced, perf or ultra)"), key)
	}
	if err := setTomlValues(root, p.values); err != nil {
		return err
	}
	c := loadConfig(root)
	c["preset"] = key
	return c.save(root)
}

func restoreToml(root string) error {
	backup := filepath.Join(root, tomlBackup)
	if !exists(backup) {
		return nil
	}
	b, err := os.ReadFile(backup)
	if err != nil {
		return err
	}
	path := filepath.Join(root, tomlName)
	if string(b) == tomlNoOrigTag {
		os.Remove(path)
	} else if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	os.Remove(backup)
	c := loadConfig(root)
	delete(c, "preset")
	delete(c, "fov")
	delete(c, "ultrawide")
	return c.save(root)
}
