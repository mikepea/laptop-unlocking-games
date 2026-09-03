package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileYieldsFreshProfile(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "nested", "profile.json"))

	p, err := store.Load("kiddo")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Name != "kiddo" {
		t.Fatalf("Name = %q, want %q", p.Name, "kiddo")
	}
	if p.PointsEarned != 0 || p.Balance() != 0 {
		t.Fatalf("fresh profile has points: earned %d, balance %d", p.PointsEarned, p.Balance())
	}
	// Maps must be usable without a nil check.
	p.Stats("typing").Plays++
	p.GrantAchievement("first-lesson")
	p.Unlock("typing")
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	store := NewStore(path)

	p := New("kiddo")
	p.AwardPoints(120)
	p.Stats("typing").Plays = 3
	p.Stats("typing").BestWPM = 27.5
	p.Stats("typing").LessonsCleared = 2
	p.GrantAchievement("first-lesson")
	p.Unlock("typing")

	if err := store.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Load("someone-else")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Name != "kiddo" {
		t.Errorf("Name = %q, want the saved name", got.Name)
	}
	if got.PointsEarned != 120 {
		t.Errorf("PointsEarned = %d, want 120", got.PointsEarned)
	}
	if s := got.Stats("typing"); s.Plays != 3 || s.BestWPM != 27.5 || s.LessonsCleared != 2 {
		t.Errorf("typing stats round-tripped as %+v", s)
	}
	if !got.HasAchievement("first-lesson") {
		t.Error("achievement did not survive the round trip")
	}
	if !got.IsUnlocked("typing") {
		t.Error("unlock did not survive the round trip")
	}
}

func TestSaveLeavesNoTempFilesBehind(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "profile.json"))

	for i := 0; i < 3; i++ {
		if err := store.Save(New("kiddo")); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "profile.json" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("directory holds %v, want just profile.json", names)
	}
}

func TestLoadRejectsNewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(path, []byte(`{"schema_version": 99}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := NewStore(path).Load("kiddo"); err == nil {
		t.Fatal("Load accepted a profile from a newer schema")
	}
}

func TestLoadReportsCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(path, []byte("not json at all"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := NewStore(path).Load("kiddo"); err == nil {
		t.Fatal("Load accepted a corrupt profile")
	}
}

func TestAwardPointsIgnoresNonPositive(t *testing.T) {
	p := New("kiddo")
	p.AwardPoints(50)
	p.AwardPoints(-100)
	p.AwardPoints(0)

	if p.PointsEarned != 50 {
		t.Fatalf("PointsEarned = %d, want 50", p.PointsEarned)
	}
}

func TestGrantAndUnlockAreIdempotent(t *testing.T) {
	p := New("kiddo")

	if !p.GrantAchievement("flawless") {
		t.Fatal("first grant reported as already held")
	}
	if p.GrantAchievement("flawless") {
		t.Fatal("second grant reported as new")
	}
	if !p.Unlock("shell") {
		t.Fatal("first unlock reported as already held")
	}
	if p.Unlock("shell") {
		t.Fatal("second unlock reported as new")
	}
}
