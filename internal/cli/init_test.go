package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestNonInteractiveInitRequiresPath(t *testing.T) {
	cmd := NewRootCommand("dev")
	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("expected input error: %v", err)
	}
}

func TestInitPersistsLocalConfigAndPreservesItOnConflict(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	root := filepath.Join(t.TempDir(), "한글 Vault")
	var out, errOut bytes.Buffer
	args := []string{"init", "--vault", root, "--create"}
	for range 2 {
		if code := Run(args, "dev", &out, &errOut); code != 0 {
			t.Fatalf("init failed: %d %s", code, &errOut)
		}
	}
	configPath := filepath.Join(configRoot, "learning-wiki", "config.json")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		VaultPath string `json:"vault_path"`
	}
	if err := json.Unmarshal(before, &config); err != nil {
		t.Fatal(err)
	}
	if config.VaultPath != root {
		t.Fatalf("wrong saved path: %q", config.VaultPath)
	}
	other := filepath.Join(t.TempDir(), "other")
	if code := Run([]string{"init", "--vault", other, "--create"}, "dev", &out, &errOut); code == 0 {
		t.Fatal("existing binding silently replaced")
	}
	after, _ := os.ReadFile(configPath)
	if !bytes.Equal(before, after) {
		t.Fatal("existing config changed")
	}
	if _, err := os.Stat(other); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("created vault despite config conflict")
	}
}

func TestInitInvalidPathAndMissingPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out, errOut bytes.Buffer
	if code := Run([]string{"init", "--vault", "relative"}, "dev", &out, &errOut); code != 2 {
		t.Fatalf("invalid input exit=%d", code)
	}
	missing := filepath.Join(t.TempDir(), "missing")
	if code := Run([]string{"init", "--vault", missing}, "dev", &out, &errOut); code != 1 {
		t.Fatalf("missing path exit=%d", code)
	}
}
