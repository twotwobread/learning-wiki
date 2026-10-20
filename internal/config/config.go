// Package config stores machine-local bindings, never secrets or vault content.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var ErrConflict = errors.New("existing configuration conflicts with requested vault")

type Config struct {
	Version   int    `json:"version"`
	VaultPath string `json:"vault_path"`
}

func Path() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(root) {
		return "", errors.New("user config directory must be absolute")
	}
	return filepath.Join(root, "learning-wiki", "config.json"), nil
}

// Check preserves unknown or corrupt existing configuration rather than overwriting it.
func Check(path, vaultPath string) error {
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%w: configuration must be a regular file", ErrConflict)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var current Config
	if err = json.Unmarshal(data, &current); err != nil {
		return fmt.Errorf("read existing configuration: %w", err)
	}
	if current.Version != 1 || current.VaultPath != vaultPath {
		return ErrConflict
	}
	return nil
}

// Save publishes a complete file without replacing an existing binding.
func Save(path, vaultPath string) (err error) {
	if err = Check(path, vaultPath); err != nil {
		return err
	}
	if _, err = os.Lstat(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".config-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.Remove(f.Name())) }()
	encodeErr := json.NewEncoder(f).Encode(Config{Version: 1, VaultPath: vaultPath})
	if err = errors.Join(encodeErr, f.Close()); err != nil {
		return err
	}
	// A same-directory hard link publishes the file without replacing an existing binding.
	if err = os.Link(f.Name(), path); errors.Is(err, fs.ErrExist) {
		return Check(path, vaultPath)
	}
	return err
}
