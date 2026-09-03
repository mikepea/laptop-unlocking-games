package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"v1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.10.0", "1.9.0", 1},
		{"2.0.0", "1.99.99", 1},
		{"1.0", "1.0.0", 0},
		{"1.0.0-dev", "1.0.0", -1},
		{"1.0.0", "1.0.0-dev", 1},
		{"0.2.0", "0.1.0-dev", 1},
		{"0.1.0", "0.1.0-dev", 1},
	}
	for _, tc := range tests {
		if got := Compare(tc.a, tc.b); got != tc.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCheckerDisabledWithoutManifestURL(t *testing.T) {
	c := &Checker{Current: "0.1.0"}
	if c.Enabled() {
		t.Fatal("Checker with no manifest URL reported enabled")
	}
	got, err := c.Check(context.Background())
	if err != nil || got != nil {
		t.Fatalf("Check = (%v, %v), want (nil, nil)", got, err)
	}
}

// manifestServer serves m at /manifest.json and the artifact bytes at /unlock.
func manifestServer(t *testing.T, version string, payload []byte) (*httptest.Server, string) {
	t.Helper()

	sum := sha256.Sum256(payload)
	hexSum := hex.EncodeToString(sum[:])

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mux.HandleFunc("/unlock", func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	})
	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Manifest{
			Version: version,
			Notes:   "a new game",
			Artifacts: map[string]Artifact{
				"linux/amd64": {URL: srv.URL + "/unlock", SHA256: hexSum, Size: int64(len(payload))},
			},
		})
	})
	return srv, hexSum
}

func TestCheckFindsNewerRelease(t *testing.T) {
	srv, _ := manifestServer(t, "0.2.0", []byte("new binary"))

	c := &Checker{ManifestURL: srv.URL + "/manifest.json", Current: "0.1.0", Platform: "linux/amd64"}
	got, err := c.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if got == nil {
		t.Fatal("Check found nothing, want 0.2.0")
	}
	if got.Version != "0.2.0" {
		t.Fatalf("Version = %q, want 0.2.0", got.Version)
	}
}

func TestCheckIgnoresOlderAndEqualReleases(t *testing.T) {
	srv, _ := manifestServer(t, "0.2.0", []byte("new binary"))
	manifestURL := srv.URL + "/manifest.json"

	for _, current := range []string{"0.2.0", "0.3.0"} {
		c := &Checker{ManifestURL: manifestURL, Current: current, Platform: "linux/amd64"}
		got, err := c.Check(context.Background())
		if err != nil {
			t.Fatalf("Check at %s: %v", current, err)
		}
		if got != nil {
			t.Fatalf("Check at %s offered %s", current, got.Version)
		}
	}
}

func TestCheckRejectsMissingPlatform(t *testing.T) {
	srv, _ := manifestServer(t, "0.2.0", []byte("new binary"))

	c := &Checker{ManifestURL: srv.URL + "/manifest.json", Current: "0.1.0", Platform: "plan9/mips"}
	if _, err := c.Check(context.Background()); err == nil {
		t.Fatal("Check accepted a release with no build for this platform")
	}
}

func TestApplyVerifiesChecksumAndReplacesTarget(t *testing.T) {
	payload := []byte("#!/bin/sh\necho new\n")
	srv, hexSum := manifestServer(t, "0.2.0", payload)

	target := filepath.Join(t.TempDir(), "unlock")
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	available := &Available{
		Version:  "0.2.0",
		Artifact: Artifact{URL: srv.URL + "/unlock", SHA256: hexSum, Size: int64(len(payload))},
	}
	if err := Apply(context.Background(), available, target, srv.Client()); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("target holds %q, want the downloaded payload", got)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("mode = %v, want 0755", info.Mode().Perm())
	}
}

func TestApplyRefusesBadChecksumAndLeavesTargetAlone(t *testing.T) {
	srv, _ := manifestServer(t, "0.2.0", []byte("new binary"))

	dir := t.TempDir()
	target := filepath.Join(dir, "unlock")
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	available := &Available{
		Version: "0.2.0",
		Artifact: Artifact{
			URL:    srv.URL + "/unlock",
			SHA256: "0000000000000000000000000000000000000000000000000000000000000000",
		},
	}
	if err := Apply(context.Background(), available, target, srv.Client()); err == nil {
		t.Fatal("Apply installed a binary whose checksum did not match")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old binary" {
		t.Fatalf("target was modified despite the failure: %q", got)
	}

	// The half-downloaded file must not be left lying next to the target.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("directory holds %d entries after a failed update, want 1", len(entries))
	}
}

func TestCheckRejectsNonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &Checker{ManifestURL: srv.URL, Current: "0.1.0", Platform: "linux/amd64"}
	if _, err := c.Check(context.Background()); err == nil {
		t.Fatal("Check accepted a 500 response")
	}
}
