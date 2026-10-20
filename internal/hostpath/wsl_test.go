package hostpath

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWSLPathPassesOneUninterpretedArgument(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture requires a Unix host")
	}
	bin := t.TempDir()
	argsFile := filepath.Join(bin, "args")
	script := "#!/bin/sh\nprintf '%s\\n' \"$#\" \"$1\" \"$2\" > \"$ARG_FILE\"\nprintf '/mnt/d/Notes\\n'\n"
	if err := os.WriteFile(filepath.Join(bin, "wslpath"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("ARG_FILE", argsFile)
	unwanted := filepath.Join(t.TempDir(), "unwanted")
	input := `D:\내 문서\notes;$(touch ` + unwanted + `)`
	got, err := WSLPath(context.Background(), input)
	if err != nil || got != "/mnt/d/Notes\n" {
		t.Fatalf("conversion returned %q, %v", got, err)
	}
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "2\n-u\n"+input+"\n" {
		t.Fatalf("argument boundary changed: %q", data)
	}
	if _, err := os.Stat(unwanted); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("path input executed as shell code")
	}
}

func TestWSLPathPreservesCommandFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture requires a Unix host")
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "wslpath"), []byte("#!/bin/sh\nexit 7\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	_, err := WSLPath(context.Background(), `D:\Notes`)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 7 {
		t.Fatalf("conversion failure lost its cause: %v", err)
	}
}
