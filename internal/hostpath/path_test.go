package hostpath

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestNativePaths(t *testing.T) {
	resolver := Native()
	want := filepath.Join(t.TempDir(), "한글 notes")
	got, err := resolver.Resolve(context.Background(), want)
	if err != nil || got != want {
		t.Fatalf("got %q, %v; want %q", got, err, want)
	}
	for _, input := range []string{"", "relative/path", `D:notes`, `D:\notes`, `\\server\share`, "//server/share", "bad\npath"} {
		_, err := resolver.Resolve(context.Background(), input)
		if !errors.Is(err, ErrInvalidPath) {
			t.Fatalf("%q: expected invalid path, got %v", input, err)
		}
	}
}

func TestWSLWindowsPathConversion(t *testing.T) {
	input := `D:\내 문서\notes;echo bad`
	want := filepath.Join(t.TempDir(), "한글 notes")
	resolver := WSL(func(_ context.Context, value string) (string, error) {
		if value != input {
			t.Fatalf("input changed: %q", value)
		}
		return want + "\n", nil
	})
	got, err := resolver.Resolve(context.Background(), input)
	if err != nil || got != want {
		t.Fatalf("got %q, %v; want %q", got, err, want)
	}
	cause := errors.New("converter unavailable")
	_, err = WSL(func(context.Context, string) (string, error) { return "", cause }).Resolve(context.Background(), input)
	if !errors.Is(err, cause) {
		t.Fatalf("lost conversion cause: %v", err)
	}
	_, err = WSL(func(context.Context, string) (string, error) { return "", nil }).Resolve(context.Background(), input)
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("empty conversion accepted: %v", err)
	}
}
