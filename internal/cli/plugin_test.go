package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/twotwobread/learning-wiki/internal/config"
	"github.com/twotwobread/learning-wiki/internal/obsidian"
)

func TestInitWithoutPluginConsentDoesNotDownload(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	root := filepath.Join(t.TempDir(), "vault")
	cmd := newInitCommandWithFetcher(func(context.Context, obsidian.Asset) ([]byte, error) {
		t.Fatal("download without consent")
		return nil, nil
	})
	cmd.SetArgs([]string{"--vault", root, "--create"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".obsidian")); !os.IsNotExist(err) {
		t.Fatal("modified Obsidian without consent")
	}
}

func TestPluginDownloadFailureDoesNotInitialize(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	root := filepath.Join(t.TempDir(), "vault")
	configPath, err := config.Path()
	if err != nil {
		t.Fatal(err)
	}
	cause := errors.New("offline")
	cmd := newInitCommandWithFetcher(func(context.Context, obsidian.Asset) ([]byte, error) { return nil, cause })
	cmd.SetArgs([]string{"--vault", root, "--create", "--plugin"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); !errors.Is(err, cause) {
		t.Fatalf("lost download error: %v", err)
	}
	for _, p := range []string{root, configPath} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("side effect before successful download: %s", p)
		}
	}
}
