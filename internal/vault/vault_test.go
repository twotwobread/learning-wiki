package vault

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestResolve(t *testing.T) {
	for _, input := range []string{"", "relative/path", `D:notes`, `\\server\share`} {
		_, err := Resolve(context.Background(), input, nil)
		if !errors.Is(err, ErrInvalidPath) {
			t.Fatalf("%q: expected invalid path, got %v", input, err)
		}
	}
	want := filepath.Join(t.TempDir(), "한글 notes")
	got, err := Resolve(context.Background(), want, nil)
	if err != nil || got != want {
		t.Fatalf("got %q, %v", got, err)
	}
	input := `D:\내 문서\notes;echo bad`
	got, err = Resolve(context.Background(), input, func(_ context.Context, value string) (string, error) {
		if value != input {
			t.Fatalf("input changed: %q", value)
		}
		return want + "\n", nil
	})
	if err != nil || got != want {
		t.Fatalf("got %q, %v", got, err)
	}
	cause := errors.New("converter unavailable")
	_, err = Resolve(context.Background(), input, func(context.Context, string) (string, error) { return "", cause })
	if !errors.Is(err, cause) {
		t.Fatalf("lost conversion cause: %v", err)
	}
	_, err = Resolve(context.Background(), input, func(context.Context, string) (string, error) { return "", nil })
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("empty conversion accepted: %v", err)
	}
}

func TestPreparePreservesDataAndIsRepeatable(t *testing.T) {
	root := t.TempDir()
	note := filepath.Join(root, "note.md")
	if err := os.WriteFile(note, []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := Prepare(root, false); err != nil {
			t.Fatal(err)
		}
	}
	content, err := os.ReadFile(note)
	if err != nil || string(content) != "keep me" {
		t.Fatalf("note changed: %q, %v", content, err)
	}
	for _, name := range []string{"sessions", "templates"} {
		info, err := os.Stat(filepath.Join(root, name))
		if err != nil || !info.IsDir() {
			t.Fatalf("directory missing: %s, %v", name, err)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("unexpected leftover files: %v", entries)
	}
}

func TestMissingDirectoryRequiresConsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new")
	if err := Prepare(path, false); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected missing path, got %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("path created without consent: %v", err)
	}
	if err := Prepare(path, true); err != nil {
		t.Fatal(err)
	}
}

func TestConflictsDoNotOverwrite(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "templates")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Prepare(root, false); !errors.Is(err, ErrNotDirectory) {
		t.Fatalf("expected conflict, got %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "keep" {
		t.Fatal("conflict overwritten")
	}
	if _, err := os.Stat(filepath.Join(root, "sessions")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("created directory before conflict preflight")
	}
	if err := Prepare(path, false); !errors.Is(err, ErrNotDirectory) {
		t.Fatalf("file accepted: %v", err)
	}
}

func TestManagedSymlinkRejected(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "sessions")); err != nil {
		t.Fatal(err)
	}
	if err := Prepare(root, false); !errors.Is(err, ErrSymlink) {
		t.Fatalf("symlink accepted: %v", err)
	}
}
