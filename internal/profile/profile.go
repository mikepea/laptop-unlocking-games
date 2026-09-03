// Package profile stores the player's persistent progress: points earned,
// per-game statistics, achievements, and which stages of the laptop they have
// unlocked. Everything lives in a single JSON file so it stays inspectable and
// easy to hand-edit when something goes wrong.
package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// schemaVersion is bumped whenever the on-disk shape changes incompatibly.
const schemaVersion = 1

// GameStats is the running record for a single game.
type GameStats struct {
	Plays          int       `json:"plays"`
	BestScore      int       `json:"best_score"`
	BestWPM        float64   `json:"best_wpm"`
	BestAccuracy   float64   `json:"best_accuracy"`
	SecondsPlayed  int       `json:"seconds_played"`
	LessonsCleared int       `json:"lessons_cleared"`
	LastPlayed     time.Time `json:"last_played"`
}

// Profile is the whole of a player's state.
type Profile struct {
	SchemaVersion int    `json:"schema_version"`
	Name          string `json:"name"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// PointsEarned only ever grows; PointsSpent tracks what has been redeemed
	// against unlocks. Balance is the difference.
	PointsEarned int `json:"points_earned"`
	PointsSpent  int `json:"points_spent"`

	Games        map[string]*GameStats `json:"games"`
	Achievements map[string]time.Time  `json:"achievements"`
	Unlocked     map[string]time.Time  `json:"unlocked"`
}

// New returns an empty profile for the named player.
func New(name string) *Profile {
	now := time.Now().UTC()
	return &Profile{
		SchemaVersion: schemaVersion,
		Name:          name,
		CreatedAt:     now,
		UpdatedAt:     now,
		Games:         map[string]*GameStats{},
		Achievements:  map[string]time.Time{},
		Unlocked:      map[string]time.Time{},
	}
}

// Balance is the points available to spend.
func (p *Profile) Balance() int { return p.PointsEarned - p.PointsSpent }

// Stats returns the record for gameID, creating an empty one if needed.
func (p *Profile) Stats(gameID string) *GameStats {
	if p.Games == nil {
		p.Games = map[string]*GameStats{}
	}
	s, ok := p.Games[gameID]
	if !ok {
		s = &GameStats{}
		p.Games[gameID] = s
	}
	return s
}

// HasAchievement reports whether the achievement has already been earned.
func (p *Profile) HasAchievement(id string) bool {
	_, ok := p.Achievements[id]
	return ok
}

// GrantAchievement records an achievement, returning false if it was already
// held so callers can avoid celebrating twice.
func (p *Profile) GrantAchievement(id string) bool {
	if p.HasAchievement(id) {
		return false
	}
	if p.Achievements == nil {
		p.Achievements = map[string]time.Time{}
	}
	p.Achievements[id] = time.Now().UTC()
	return true
}

// IsUnlocked reports whether a stage has been unlocked.
func (p *Profile) IsUnlocked(id string) bool {
	_, ok := p.Unlocked[id]
	return ok
}

// Unlock records a stage as unlocked, returning false if it already was.
func (p *Profile) Unlock(id string) bool {
	if p.IsUnlocked(id) {
		return false
	}
	if p.Unlocked == nil {
		p.Unlocked = map[string]time.Time{}
	}
	p.Unlocked[id] = time.Now().UTC()
	return true
}

// AwardPoints credits points to the player. Negative awards are ignored so a
// buggy game can never take progress away.
func (p *Profile) AwardPoints(n int) {
	if n <= 0 {
		return
	}
	p.PointsEarned += n
}

// Store reads and writes a Profile at a fixed path.
type Store struct {
	path string
}

// NewStore returns a Store backed by the file at path.
func NewStore(path string) *Store { return &Store{path: path} }

// Path is the file this Store reads and writes.
func (s *Store) Path() string { return s.path }

// DefaultPath is $XDG_DATA_HOME/unlock/profile.json, falling back to
// ~/.local/share/unlock/profile.json.
func DefaultPath() (string, error) {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate home directory: %w", err)
		}
		dir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dir, "unlock", "profile.json"), nil
}

// Load reads the profile. A missing file is not an error: it yields a fresh
// profile for name, so a first run just works.
func (s *Store) Load(name string) (*Profile, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return New(name), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read profile: %w", err)
	}

	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse profile %s: %w", s.path, err)
	}
	if p.SchemaVersion > schemaVersion {
		return nil, fmt.Errorf("profile %s was written by a newer version of unlock (schema %d, we understand %d)", s.path, p.SchemaVersion, schemaVersion)
	}
	// Maps are nil when the file predates a field; normalise so callers can
	// write into them without checking.
	if p.Games == nil {
		p.Games = map[string]*GameStats{}
	}
	if p.Achievements == nil {
		p.Achievements = map[string]time.Time{}
	}
	if p.Unlocked == nil {
		p.Unlocked = map[string]time.Time{}
	}
	if p.Name == "" {
		p.Name = name
	}
	p.SchemaVersion = schemaVersion
	return &p, nil
}

// Save writes the profile atomically, so a power cut mid-write cannot leave a
// half-written file where a child's progress used to be.
func (s *Store) Save(p *Profile) error {
	p.SchemaVersion = schemaVersion
	p.UpdatedAt = time.Now().UTC()

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".profile-*.json")
	if err != nil {
		return fmt.Errorf("create temp profile: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp profile: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp profile: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp profile: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("chmod temp profile: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replace profile: %w", err)
	}
	return nil
}
