package cli

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/twotwobread/learning-wiki/internal/config"
)

func TestInitDefaultInstallsSharedSkill(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	root := filepath.Join(t.TempDir(), "vault")
	var out, errOut bytes.Buffer
	args := []string{"init", "--vault", root, "--create"}
	before := map[string][]byte{}
	for range 2 {
		if code := Run(args, "dev", &out, &errOut); code != 0 {
			t.Fatalf("init failed: %s", &errOut)
		}
		for _, path := range []string{filepath.Join(home, ".agents/skills/learning-capture/SKILL.md"), filepath.Join(home, ".agents/skills/learning-capture/scripts/capture.mjs"), filepath.Join(home, ".agents/skills/learning-capture/binding.json"), filepath.Join(root, "templates/learning-note.md")} {
			data, err := os.ReadFile(path)
			if err != nil || len(data) == 0 {
				t.Fatalf("missing resource: %s %v", path, err)
			}
			if previous, ok := before[path]; ok && !bytes.Equal(previous, data) {
				t.Fatalf("resource changed on repeat init: %s", path)
			}
			before[path] = data
		}
	}
}

func TestInitResourceConflictHasNoVaultSideEffects(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	skill := filepath.Join(home, ".agents/skills/learning-capture/SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("user skill"), 0600); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "vault")
	var out, errOut bytes.Buffer
	if code := Run([]string{"init", "--vault", root, "--create", "--agents", "pi"}, "dev", &out, &errOut); code != 1 {
		t.Fatalf("expected conflict exit, got %d", code)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("vault created despite resource conflict")
	}
	data, _ := os.ReadFile(skill)
	if string(data) != "user skill" {
		t.Fatal("user skill changed")
	}
}

func TestInitNoSkillInstallsTemplateOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	root := filepath.Join(t.TempDir(), "vault")
	var out, errOut bytes.Buffer
	if code := Run([]string{"init", "--vault", root, "--create", "--no-skill"}, "dev", &out, &errOut); code != 0 {
		t.Fatalf("init exit=%d: %s", code, &errOut)
	}
	configPath, err := config.Path()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{configPath, filepath.Join(root, "templates/learning-note.md")} {
		if data, err := os.ReadFile(path); err != nil || len(data) == 0 {
			t.Fatalf("missing resource: %s %v", path, err)
		}
	}
	for _, path := range []string{filepath.Join(home, ".agents"), filepath.Join(root, ".obsidian")} {
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("unexpected resource: %s %v", path, err)
		}
	}
}
