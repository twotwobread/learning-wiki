package obsidian

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/twotwobread/learning-wiki/internal/resources"
)

func fixtureRelease() (Release, map[string][]byte) {
	bodies := map[string][]byte{"manifest.json": []byte(`{"id":"karpathywiki","version":"1.28.0","minAppVersion":"1.11.4"}`), "main.js": []byte("plugin code"), "styles.css": []byte("css")}
	r := Release{Version: "1.28.0"}
	for _, name := range []string{"manifest.json", "main.js", "styles.css"} {
		sum := sha256.Sum256(bodies[name])
		r.Assets = append(r.Assets, Asset{Name: name, URL: "https://example.invalid/" + name, SHA256: hex.EncodeToString(sum[:])})
	}
	return r, bodies
}

func TestPrepareAndInstallPreservesSettings(t *testing.T) {
	root := t.TempDir()
	r, bodies := fixtureRelease()
	settings := filepath.Join(root, ".obsidian", "community-plugins.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings, []byte(`["other-plugin"]`), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := Prepare(context.Background(), root, r, func(_ context.Context, a Asset) ([]byte, error) { return bodies[a.Name], nil })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".obsidian/plugins")); !os.IsNotExist(err) {
		t.Fatal("prepare wrote files")
	}
	if err := resources.Install(result.Files); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(settings); string(data) != `["other-plugin"]` {
		t.Fatal("activation settings changed")
	}
	plugin := filepath.Join(root, ".obsidian/plugins/karpathywiki")
	if err := os.WriteFile(filepath.Join(plugin, "data.json"), []byte("private configuration"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err = Prepare(context.Background(), root, r, func(context.Context, Asset) ([]byte, error) {
		t.Fatal("existing plugin triggered download")
		return nil, nil
	})
	if err != nil || !result.Existing || len(result.Files) != 0 {
		t.Fatalf("existing plugin not preserved: %+v %v", result, err)
	}
	if data, _ := os.ReadFile(filepath.Join(plugin, "data.json")); string(data) != "private configuration" {
		t.Fatal("plugin settings changed")
	}
}

func TestDownloadFailuresDoNotWrite(t *testing.T) {
	for _, kind := range []string{"network", "checksum", "manifest"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			r, bodies := fixtureRelease()
			cause := errors.New("offline")
			want := ErrIntegrity
			if kind == "network" {
				want = cause
			}
			if kind == "manifest" {
				bodies["manifest.json"] = []byte(`{"id":"wrong","version":"1.28.0"}`)
				sum := sha256.Sum256(bodies["manifest.json"])
				r.Assets[0].SHA256 = hex.EncodeToString(sum[:])
				want = ErrManifest
			}
			_, err := Prepare(context.Background(), root, r, func(_ context.Context, a Asset) ([]byte, error) {
				if kind == "network" {
					return nil, cause
				}
				if kind == "checksum" {
					return []byte("tampered"), nil
				}
				return bodies[a.Name], nil
			})
			if !errors.Is(err, want) {
				t.Fatalf("expected %v, got %v", want, err)
			}
			entries, _ := os.ReadDir(root)
			if len(entries) != 0 {
				t.Fatal("failed download changed vault")
			}
		})
	}
}

func TestRejectIncompleteOrSymlinkedInstallation(t *testing.T) {
	for _, link := range []bool{false, true} {
		t.Run(map[bool]string{false: "incomplete", true: "symlink"}[link], func(t *testing.T) {
			root := t.TempDir()
			r, _ := fixtureRelease()
			if link {
				if err := os.Symlink(t.TempDir(), filepath.Join(root, ".obsidian")); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.MkdirAll(filepath.Join(root, ".obsidian/plugins/karpathywiki"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			_, err := Prepare(context.Background(), root, r, func(context.Context, Asset) ([]byte, error) { t.Fatal("conflict triggered download"); return nil, nil })
			if !errors.Is(err, ErrExisting) {
				t.Fatalf("expected conflict: %v", err)
			}
		})
	}
}

func TestPrepareAllowsSymlinkOutsideVault(t *testing.T) {
	parent := t.TempDir()
	real := filepath.Join(parent, "real")
	if err := os.Mkdir(real, 0755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(alias, "vault")
	release, bodies := fixtureRelease()
	result, err := Prepare(context.Background(), root, release, func(_ context.Context, a Asset) ([]byte, error) { return bodies[a.Name], nil })
	if err != nil {
		t.Fatal(err)
	}
	if err := resources.Install(result.Files); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".obsidian", "plugins", PluginID, "manifest.json")); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareAcceptsTrailingSeparatorInVaultPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault") + string(filepath.Separator)
	release, bodies := fixtureRelease()
	result, err := Prepare(context.Background(), root, release, func(_ context.Context, a Asset) ([]byte, error) { return bodies[a.Name], nil })
	if err != nil || len(result.Files) != len(release.Assets) {
		t.Fatalf("unexpected result: %+v %v", result, err)
	}
}
