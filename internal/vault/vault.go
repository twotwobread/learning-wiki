// Package vault handles local paths without depending on a synchronization provider.
package vault

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidPath  = errors.New("expected an absolute Windows drive or Linux path")
	ErrNotDirectory = errors.New("path is not a directory")
	ErrSymlink      = errors.New("managed directory must not be a symbolic link")
)

type Converter func(context.Context, string) (string, error)

// WSLPath passes the path as one argument, never through a shell.
func WSLPath(ctx context.Context, path string) (string, error) {
	output, err := exec.CommandContext(ctx, "wslpath", "-u", path).Output()
	if err != nil {
		return "", fmt.Errorf("convert Windows path: %w", err)
	}
	return string(output), nil
}

func Resolve(ctx context.Context, input string, convert Converter) (string, error) {
	if strings.TrimSpace(input) == "" || strings.ContainsAny(input, "\x00\r\n") || strings.HasPrefix(input, "//") || strings.HasPrefix(input, `\\`) {
		return "", ErrInvalidPath
	}
	windows := len(input) >= 3 && ((input[0] >= 'A' && input[0] <= 'Z') || (input[0] >= 'a' && input[0] <= 'z')) && input[1] == ':' && (input[2] == '\\' || input[2] == '/')
	if windows {
		if convert == nil {
			convert = WSLPath
		}
		output, err := convert(ctx, input)
		if err != nil {
			return "", err
		}
		input = strings.TrimSuffix(strings.TrimSuffix(output, "\n"), "\r")
	}
	if !filepath.IsAbs(input) || strings.ContainsAny(input, "\x00\r\n") {
		return "", ErrInvalidPath
	}
	return filepath.Clean(input), nil
}

func checkDir(path string, rejectSymlink bool) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		if rejectSymlink {
			return fmt.Errorf("%w: %s", ErrSymlink, path)
		}
		info, err = os.Stat(path)
		if err != nil {
			return err
		}
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %s", ErrNotDirectory, path)
	}
	return nil
}

// CheckAccess probes local read/write access and removes its own temporary file.
func CheckAccess(root string) (err error) {
	if err = checkDir(root, false); err != nil {
		return err
	}
	if _, err = os.ReadDir(root); err != nil {
		return err
	}
	f, err := os.CreateTemp(root, ".learning-wiki-check-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.Remove(f.Name())) }()
	_, writeErr := f.WriteString("learning-wiki access check")
	if err = errors.Join(writeErr, f.Close()); err != nil {
		return err
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		return err
	}
	if string(data) != "learning-wiki access check" {
		return errors.New("local file readback mismatch")
	}
	return nil
}

// Prepare never changes existing files. New directories require explicit root creation consent.
// Preflight catches existing conflicts; on failure only newly created empty directories are removed.
func Prepare(root string, create bool) (err error) {
	if !filepath.IsAbs(root) {
		return ErrInvalidPath
	}
	root = filepath.Clean(root)
	paths := []string{root, filepath.Join(root, "sessions"), filepath.Join(root, "templates")}
	for i, path := range paths {
		e := checkDir(path, i > 0)
		if errors.Is(e, fs.ErrNotExist) {
			if i == 0 && !create {
				return e
			}
			continue
		}
		if e != nil {
			return e
		}
	}
	var created []string
	defer func() {
		if err != nil {
			for i := len(created) - 1; i >= 0; i-- {
				err = errors.Join(err, os.Remove(created[i]))
			}
		}
	}()
	// Track every new ancestor so failed initialization can remove only its own empty directories.
	var ensure func(string) error
	ensure = func(path string) error {
		if e := checkDir(path, false); e == nil {
			return nil
		} else if !errors.Is(e, fs.ErrNotExist) {
			return e
		}
		if e := ensure(filepath.Dir(path)); e != nil {
			return e
		}
		if e := os.Mkdir(path, 0755); e != nil {
			return e
		}
		created = append(created, path)
		return nil
	}
	if err = ensure(root); err != nil {
		return err
	}
	if err = CheckAccess(root); err != nil {
		return err
	}
	for _, path := range paths[1:] {
		if err = ensure(path); err != nil {
			return err
		}
		if err = checkDir(path, true); err != nil {
			return err
		}
		if err = CheckAccess(path); err != nil {
			return err
		}
	}
	return nil
}
