// Package resources installs embedded templates and portable Agent Skills without overwriting files.
package resources

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed assets/*
var assets embed.FS
var ErrConflict = errors.New("resource conflicts with an existing file or symlink")
var ErrInvalidAgent = errors.New("supported agents: pi, codex")

type File struct {
	Path    string
	Content []byte
}

func ValidateAgents(agents []string) error {
	for _, a := range agents {
		if a != "pi" && a != "codex" {
			return fmt.Errorf("%w: %s", ErrInvalidAgent, a)
		}
	}
	return nil
}
func Plan(home, root, configPath string, agents []string) ([]File, error) {
	if err := ValidateAgents(agents); err != nil {
		return nil, err
	}
	template, err := assets.ReadFile("assets/learning-note.md")
	if err != nil {
		return nil, err
	}
	files := []File{{filepath.Join(root, "templates", "learning-note.md"), template}}
	if len(agents) == 0 {
		return files, nil
	}
	skill := filepath.Join(home, ".agents", "skills", "learning-capture")
	for _, entry := range []struct{ src, dest string }{{"SKILL.md", "SKILL.md"}, {"capture.mjs", "scripts/capture.mjs"}} {
		data, err := assets.ReadFile("assets/" + entry.src)
		if err != nil {
			return nil, err
		}
		files = append(files, File{filepath.Join(skill, entry.dest), data})
	}
	data, err := json.MarshalIndent(struct {
		ConfigPath string `json:"config_path"`
	}{configPath}, "", "  ")
	if err != nil {
		return nil, err
	}
	files = append(files, File{filepath.Join(skill, "binding.json"), append(data, '\n')})
	return files, nil
}

// Check runs without creating any files or directories.
func Check(files []File) error {
	for _, f := range files {
		if !filepath.IsAbs(f.Path) {
			return fmt.Errorf("resource path must be absolute: %s", f.Path)
		}
		for p := filepath.Dir(f.Path); ; p = filepath.Dir(p) {
			info, err := os.Lstat(p)
			if err == nil && !info.IsDir() {
				return fmt.Errorf("%w: %s", ErrConflict, p)
			}
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if filepath.Dir(p) == p {
				break
			}
		}
		info, err := os.Lstat(f.Path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%w: %s", ErrConflict, f.Path)
		}
		data, err := os.ReadFile(f.Path)
		if err != nil {
			return err
		}
		if !bytes.Equal(data, f.Content) {
			return fmt.Errorf("%w: %s", ErrConflict, f.Path)
		}
	}
	return nil
}

func Install(files []File) (err error) {
	if err = Check(files); err != nil {
		return err
	}
	var created []string
	defer func() {
		if err != nil {
			for _, p := range created {
				err = errors.Join(err, os.Remove(p))
			}
		}
	}()
	for _, f := range files {
		if err = Check([]File{f}); err != nil {
			return err
		}
		if _, e := os.Lstat(f.Path); e == nil {
			continue
		} else if !errors.Is(e, os.ErrNotExist) {
			return e
		}
		if err = os.MkdirAll(filepath.Dir(f.Path), 0755); err != nil {
			return err
		}
		// Exclusive creation prevents replacing files created after the preflight.
		out, e := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return e
		}
		created = append(created, f.Path)
		_, e = out.Write(f.Content)
		err = errors.Join(e, out.Close())
		if err != nil {
			return err
		}
	}
	return nil
}
