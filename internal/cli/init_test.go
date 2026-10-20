package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/twotwobread/learning-wiki/internal/config"
	"github.com/twotwobread/learning-wiki/internal/hostpath"
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

func TestInitUsesSelectedPathResolver(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	configPath, err := config.Path()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "vault")
	windows := `D:\Notes\Vault`
	resolver := hostpath.WSL(func(_ context.Context, input string) (string, error) {
		if input != windows {
			t.Fatalf("unexpected path input: %q", input)
		}
		return root, nil
	})
	cmd := newRootCommandWithResolver("dev", resolver)
	cmd.SetArgs([]string{"init", "--vault", windows, "--create"})
	cmd.SetOut(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		VaultPath string `json:"vault_path"`
	}
	if err := json.Unmarshal(data, &saved); err != nil || saved.VaultPath != root {
		t.Fatalf("wrong saved path: %+v %v", saved, err)
	}
}

func TestInitPersistsLocalConfigAndPreservesItOnConflict(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	configPath, err := config.Path()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "한글 Vault")
	var out, errOut bytes.Buffer
	args := []string{"init", "--vault", root, "--create"}
	for range 2 {
		if code := Run(args, "dev", &out, &errOut); code != 0 {
			t.Fatalf("init failed: %d %s", code, &errOut)
		}
	}
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
	t.Setenv("HOME", t.TempDir())
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

func TestInitNoSkillRejectsExplicitAgentsWithoutSideEffects(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	root := filepath.Join(t.TempDir(), "vault")
	configPath, err := config.Path()
	if err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := Run([]string{"init", "--vault", root, "--create", "--no-skill", "--agents", "codex"}, "dev", &out, &errOut); code != 2 {
		t.Fatalf("invalid input exit=%d", code)
	}
	for _, path := range []string{root, configPath, filepath.Join(home, ".agents")} {
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("created resource despite invalid arguments: %s %v", path, err)
		}
	}
}

func TestInitNoSkillRequiresVault(t *testing.T) {
	cmd := NewRootCommand("dev")
	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"init", "--no-skill"})
	if err := cmd.Execute(); !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("expected input error: %v", err)
	}
}
