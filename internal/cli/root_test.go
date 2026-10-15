package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestInformationalCommands(t *testing.T) {
	for _, args := range [][]string{{}, {"--help"}, {"-h"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, errOut bytes.Buffer
			cmd := NewRootCommand("test-version")
			cmd.SetArgs(args)
			cmd.SetOut(&out)
			cmd.SetErr(&errOut)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.Len() == 0 {
				t.Fatal("expected help output")
			}
			if errOut.Len() != 0 {
				t.Fatalf("unexpected diagnostic: %s", &errOut)
			}
		})
	}
}

func TestVersionReportsInjectedValue(t *testing.T) {
	for _, version := range []string{"dev", "0.1.0", "0.2.0-rc.1+build123"} {
		t.Run(version, func(t *testing.T) {
			var out bytes.Buffer
			cmd := NewRootCommand(version)
			cmd.SetArgs([]string{"--version"})
			cmd.SetOut(&out)
			cmd.SetErr(io.Discard)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Version is semantic output; surrounding presentation is not a contract.
			if !strings.Contains(out.String(), version) {
				t.Fatalf("missing version %q in %q", version, out.String())
			}
		})
	}
}

func TestInvalidInputHasSemanticError(t *testing.T) {
	for _, args := range [][]string{{"unknown-command"}, {"--unknown-option"}, {"extra", "arguments"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := NewRootCommand("dev")
			cmd.SetArgs(args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			if err := cmd.Execute(); !errors.Is(err, ErrInvalidArguments) {
				t.Fatalf("expected ErrInvalidArguments, got %v", err)
			}
		})
	}
}

func TestRunExitStatusAndOutputRouting(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		wantCode int
	}{
		{"help", []string{"--help"}, 0},
		{"version", []string{"--version"}, 0},
		{"no arguments", nil, 0},
		{"invalid input", []string{"--unknown"}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := Run(tc.args, "dev", &out, &errOut)
			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d", code, tc.wantCode)
			}
			if code == 0 && (out.Len() == 0 || errOut.Len() != 0) {
				t.Fatal("expected informational stdout without stderr")
			}
			if code != 0 && (errOut.Len() == 0 || out.Len() != 0) {
				t.Fatal("expected failure diagnostic on stderr only")
			}
		})
	}
}
