package config

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveRoundTripAndConflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	root := t.TempDir()
	if err := Save(path, root); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(path, root); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, t.TempDir()); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict: %v", err)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("existing config changed")
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temporary files left: %v", entries)
	}
}

func TestCorruptConfigPreserved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := []byte("not valid JSON")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, t.TempDir()); err == nil {
		t.Fatal("corrupt config accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(original, after) {
		t.Fatal("corrupt config overwritten")
	}
}
