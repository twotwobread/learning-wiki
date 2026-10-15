// Package obsidian installs upstream distribution files, never generated wiki content or credentials.
package obsidian

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/twotwobread/learning-wiki/internal/resources"
)

const PluginID = "karpathywiki"
const Version = "1.28.0"
const MinAppVersion = "1.11.4"
const Repository = "https://github.com/GD4AI/obsidian-llm-wiki"
const maxAssetBytes = 16 << 20

var (
	ErrIntegrity = errors.New("plugin asset checksum mismatch")
	ErrManifest  = errors.New("invalid plugin manifest")
	ErrExisting  = errors.New("existing plugin installation requires manual review")
)

type Asset struct{ Name, URL, SHA256 string }
type Release struct {
	Version string
	Assets  []Asset
}
type Fetcher func(context.Context, Asset) ([]byte, error)
type Result struct {
	Existing bool
	Version  string
	Files    []resources.File
}
type manifest struct {
	ID            string `json:"id"`
	Version       string `json:"version"`
	MinAppVersion string `json:"minAppVersion"`
}

// PinnedRelease hashes were checked against GitHub release metadata (runtime files)
// and the tagged upstream source (license notices). Updates require explicit review.
func PinnedRelease() Release {
	base := Repository + "/releases/download/" + Version + "/"
	raw := "https://raw.githubusercontent.com/GD4AI/obsidian-llm-wiki/" + Version + "/"
	return Release{Version: Version, Assets: []Asset{
		{"manifest.json", base + "manifest.json", "7469ab62415add60f89ae6073ed0fbea2c7cdb0d22057323a500666826492500"},
		{"main.js", base + "main.js", "849bb6790d39091dd1446901d1f1c70f489bfd57552c2b71679fdf7aeee4be2e"},
		{"styles.css", base + "styles.css", "23f9bdc6dbe57ccf92b4d96818e69086f6ede34657283248e7175a2d3350a47c"},
		{"LICENSE", raw + "LICENSE", "b669d54594b161f62cb9f6d4f23ff74f8431d50ecd63b6d101913402087cb081"},
		{"NOTICE", raw + "NOTICE", "4caa919dd397d9cdce955407934f3cf3ef6f12d3e7d520bdea57496d7e217c4f"},
		{"THIRD-PARTY-NOTICES.md", raw + "THIRD-PARTY-NOTICES.md", "93a65660afab7379ad484c5fc6946d4a514b5b67ae6367bd14e428049a840761"},
	}}
}

func Fetch(ctx context.Context, a Asset) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if err != nil {
		return nil, err
	}
	if req.URL.Scheme != "https" {
		return nil, errors.New("plugin download requires HTTPS")
	}
	client := &http.Client{Timeout: 90 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" || len(via) >= 10 {
			return errors.New("unsafe or excessive redirect")
		}
		return nil
	}}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plugin download HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxAssetBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxAssetBytes {
		return nil, errors.New("plugin asset exceeds size limit")
	}
	return data, nil
}

// Prepare downloads and verifies in memory only. Existing valid installations are
// preserved, including other versions; it neither upgrades nor validates their provenance.
func Prepare(ctx context.Context, root string, release Release, fetch Fetcher) (Result, error) {
	if !filepath.IsAbs(root) {
		return Result{}, errors.New("vault path must be absolute")
	}
	dest := filepath.Join(root, ".obsidian", "plugins", PluginID)
	for p := dest; ; p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if err == nil && !info.IsDir() {
			return Result{}, fmt.Errorf("%w: %s", ErrExisting, p)
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return Result{}, err
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	if _, err := os.Lstat(dest); err == nil {
		for _, name := range []string{"manifest.json", "main.js", "styles.css"} {
			info, err := os.Lstat(filepath.Join(dest, name))
			if err != nil || !info.Mode().IsRegular() {
				return Result{}, fmt.Errorf("%w: missing or unsafe %s", ErrExisting, name)
			}
		}
		data, err := os.ReadFile(filepath.Join(dest, "manifest.json"))
		if err != nil {
			return Result{}, err
		}
		var m manifest
		if json.Unmarshal(data, &m) != nil || m.ID != PluginID || m.Version == "" {
			return Result{}, ErrExisting
		}
		return Result{Existing: true, Version: m.Version}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return Result{}, err
	}
	if fetch == nil {
		fetch = Fetch
	}
	result := Result{Version: release.Version}
	seen := map[string]bool{}
	for _, a := range release.Assets {
		if filepath.Base(a.Name) != a.Name || a.Name == "." || a.Name == ".." || seen[a.Name] {
			return Result{}, ErrManifest
		}
		seen[a.Name] = true
		data, err := fetch(ctx, a)
		if err != nil {
			return Result{}, fmt.Errorf("download %s: %w", a.Name, err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != a.SHA256 {
			return Result{}, fmt.Errorf("%w: %s", ErrIntegrity, a.Name)
		}
		if a.Name == "manifest.json" {
			var m manifest
			if json.Unmarshal(data, &m) != nil || m.ID != PluginID || m.Version != release.Version || m.MinAppVersion != MinAppVersion {
				return Result{}, ErrManifest
			}
		}
		result.Files = append(result.Files, resources.File{Path: filepath.Join(dest, a.Name), Content: data})
	}
	for _, name := range []string{"manifest.json", "main.js", "styles.css"} {
		if !seen[name] {
			return Result{}, ErrManifest
		}
	}
	if err := resources.Check(result.Files); err != nil {
		return Result{}, err
	}
	return result, nil
}
