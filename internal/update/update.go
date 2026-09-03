// Package update fetches a signed-by-checksum manifest, works out whether a
// newer build exists for this platform, and swaps the binary in place.
//
// The design assumption is that the laptop is offline as often as not, so every
// failure path here is non-fatal: the launcher checks in the background and
// simply says nothing when the check does not work out.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Artifact is one downloadable build.
type Artifact struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Manifest is the document published alongside a release.
type Manifest struct {
	Version string `json:"version"`
	Notes   string `json:"notes"`
	// Artifacts is keyed by "GOOS/GOARCH", e.g. "linux/amd64".
	Artifacts map[string]Artifact `json:"artifacts"`
}

// Available describes an update that applies to this machine.
type Available struct {
	Version  string
	Notes    string
	Artifact Artifact
}

// Checker asks a manifest URL whether there is anything newer.
type Checker struct {
	// ManifestURL is where the release manifest lives. Empty disables updates.
	ManifestURL string
	// Current is the version this binary reports.
	Current string
	// Platform defaults to runtime GOOS/GOARCH.
	Platform string
	// Client defaults to a client with a short timeout.
	Client *http.Client
}

// Enabled reports whether this Checker will do anything.
func (c *Checker) Enabled() bool { return c.ManifestURL != "" }

func (c *Checker) platform() string {
	if c.Platform != "" {
		return c.Platform
	}
	return runtime.GOOS + "/" + runtime.GOARCH
}

func (c *Checker) client() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return &http.Client{Timeout: 15 * time.Second}
}

// Check returns the update that applies, or nil when the build is current.
func (c *Checker) Check(ctx context.Context) (*Available, error) {
	if !c.Enabled() {
		return nil, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.ManifestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build manifest request: %w", err)
	}
	resp, err := c.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch manifest: unexpected status %s", resp.Status)
	}

	var m Manifest
	// Cap the read: a manifest is a few hundred bytes, and this runs
	// unattended against whatever the network hands back.
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if Compare(m.Version, c.Current) <= 0 {
		return nil, nil
	}
	art, ok := m.Artifacts[c.platform()]
	if !ok {
		return nil, fmt.Errorf("release %s has no build for %s", m.Version, c.platform())
	}
	if art.URL == "" || art.SHA256 == "" {
		return nil, fmt.Errorf("release %s for %s is missing a url or checksum", m.Version, c.platform())
	}
	return &Available{Version: m.Version, Notes: m.Notes, Artifact: art}, nil
}

// Apply downloads the update, verifies its checksum, and replaces the file at
// targetPath. The download is written next to the target so the final rename is
// atomic and stays on one filesystem.
//
// On Linux, replacing the file a running process was started from is safe: the
// running image keeps the old inode until it exits.
func Apply(ctx context.Context, u *Available, targetPath string, client *http.Client) error {
	if u == nil {
		return nil
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Minute}
	}

	target, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", targetPath, err)
	}
	dir := filepath.Dir(target)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.Artifact.URL, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", u.Artifact.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected status %s", u.Artifact.URL, resp.Status)
	}

	tmp, err := os.CreateTemp(dir, ".aido-update-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds

	sum := sha256.New()
	body := io.Reader(resp.Body)
	if u.Artifact.Size > 0 {
		// Refuse to be fed an unbounded stream by a hostile or broken mirror.
		body = io.LimitReader(body, u.Artifact.Size)
	}
	written, err := io.Copy(io.MultiWriter(tmp, sum), body)
	if err != nil {
		tmp.Close()
		return fmt.Errorf("write update: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close update: %w", err)
	}
	if u.Artifact.Size > 0 && written != u.Artifact.Size {
		return fmt.Errorf("update truncated: got %d bytes, expected %d", written, u.Artifact.Size)
	}

	got := hex.EncodeToString(sum.Sum(nil))
	if !strings.EqualFold(got, u.Artifact.SHA256) {
		return fmt.Errorf("checksum mismatch for %s: got %s, expected %s", u.Artifact.URL, got, u.Artifact.SHA256)
	}

	if err := os.Chmod(tmpName, 0o755); err != nil {
		return fmt.Errorf("chmod update: %w", err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("install update over %s: %w", target, err)
	}
	return nil
}

// Compare orders two dotted numeric versions, ignoring a leading "v" and any
// pre-release suffix after "-". It returns -1, 0 or 1 for a < b, a == b, a > b.
//
// This is deliberately smaller than full semver: releases here are plain
// x.y.z tags, and a pre-release is always treated as older than the release it
// leads to.
func Compare(a, b string) int {
	aNums, aPre := splitVersion(a)
	bNums, bPre := splitVersion(b)

	for i := 0; i < len(aNums) || i < len(bNums); i++ {
		av, bv := 0, 0
		if i < len(aNums) {
			av = aNums[i]
		}
		if i < len(bNums) {
			bv = bNums[i]
		}
		if av != bv {
			if av < bv {
				return -1
			}
			return 1
		}
	}

	switch {
	case aPre == bPre:
		return 0
	case aPre != "" && bPre == "":
		return -1 // 1.0.0-dev is older than 1.0.0
	case aPre == "" && bPre != "":
		return 1
	case aPre < bPre:
		return -1
	default:
		return 1
	}
}

func splitVersion(v string) (nums []int, pre string) {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		pre = v[i+1:]
		v = v[:i]
	}
	for _, part := range strings.Split(v, ".") {
		n, err := strconv.Atoi(part)
		if err != nil {
			// Anything unparseable sorts as 0 rather than failing the whole
			// comparison; an update check must never crash the launcher.
			n = 0
		}
		nums = append(nums, n)
	}
	return nums, pre
}
