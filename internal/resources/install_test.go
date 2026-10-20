package resources

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallRepeatableAndShared(t *testing.T) {
	home, root := t.TempDir(), t.TempDir()
	files, err := Plan(home, root, "/local/config.json", []string{"pi", "codex"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := Install(files); err != nil {
			t.Fatal(err)
		}
	}
	skill := filepath.Join(home, ".agents", "skills", "learning-capture", "SKILL.md")
	if data, err := os.ReadFile(skill); err != nil || len(data) == 0 {
		t.Fatalf("skill unavailable: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "templates", "learning-note.md")); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Dir(filepath.Dir(skill)))
	if len(entries) != 1 {
		t.Fatal("duplicate skill installation")
	}
}

func TestConflictPreflightPreservesAllFiles(t *testing.T) {
	home, root := t.TempDir(), t.TempDir()
	files, err := Plan(home, root, "/config.json", []string{"pi"})
	if err != nil {
		t.Fatal(err)
	}
	conflict := files[len(files)-1]
	if err := os.MkdirAll(filepath.Dir(conflict.Path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(conflict.Path, []byte("user edit"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Install(files); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict: %v", err)
	}
	data, _ := os.ReadFile(conflict.Path)
	if string(data) != "user edit" {
		t.Fatal("user file overwritten")
	}
	if _, err := os.Stat(files[0].Path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("wrote before preflight completed")
	}
}

func TestNoAgentAndInvalidAgent(t *testing.T) {
	files, err := Plan(t.TempDir(), t.TempDir(), "/config", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatal("no-agent selection must install only template")
	}
	if _, err := Plan(t.TempDir(), t.TempDir(), "/config", []string{"unknown"}); !errors.Is(err, ErrInvalidAgent) {
		t.Fatalf("invalid agent accepted: %v", err)
	}
}

func TestRejectSymlinkParent(t *testing.T) {
	home, root, outside := t.TempDir(), t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, ".agents")); err != nil {
		t.Fatal(err)
	}
	files, err := Plan(home, root, "/config", []string{"codex"})
	if err != nil {
		t.Fatal(err)
	}
	if err := Install(files); !errors.Is(err, ErrConflict) {
		t.Fatalf("symlink accepted: %v", err)
	}
	entries, _ := os.ReadDir(outside)
	if len(entries) != 0 {
		t.Fatal("wrote through symlink")
	}
}

func TestAllowsSymlinkOutsideManagedRoots(t *testing.T) {
	parent := t.TempDir()
	real := filepath.Join(parent, "real")
	if err := os.Mkdir(real, 0755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(alias, "home")
	root := filepath.Join(alias, "vault")
	files, err := Plan(home, root, "/config", []string{"codex"})
	if err != nil {
		t.Fatal(err)
	}
	if err := Install(files); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "templates", "learning-note.md")); err != nil {
		t.Fatal(err)
	}
}

func TestInstallAcceptsTrailingSeparatorInManagedRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault") + string(filepath.Separator)
	files, err := Plan(t.TempDir(), root, "/config", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := Install(files); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "templates", "learning-note.md")); err != nil {
		t.Fatal(err)
	}
}
