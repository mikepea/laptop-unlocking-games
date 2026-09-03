package launcher

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
	"github.com/mikepea/laptop-unlocking-games/internal/games/typing"
	"github.com/mikepea/laptop-unlocking-games/internal/points"
	"github.com/mikepea/laptop-unlocking-games/internal/profile"
)

func newTestModel(t *testing.T) (*Model, *profile.Store) {
	t.Helper()
	store := profile.NewStore(filepath.Join(t.TempDir(), "profile.json"))
	prof := profile.New("kiddo")
	m := New(Options{
		Store:   store,
		Profile: prof,
		Games:   games.NewRegistry(typing.New()),
		Ledger:  points.NewLocal(prof),
	})
	return m, store
}

func passResult(round string, index, pts int) games.Result {
	return games.Result{
		GameID:     typing.GameID,
		Completed:  true,
		Round:      round,
		RoundIndex: index,
		Score:      pts,
		Points:     pts,
		Accuracy:   0.95,
		WPM:        25,
		Duration:   90 * time.Second,
	}
}

func TestNewUnlocksTheFreeStage(t *testing.T) {
	m, _ := newTestModel(t)

	if !m.prof.IsUnlocked("typing") {
		t.Fatal("the zero-cost stage was not unlocked on a fresh profile")
	}
	if m.prof.IsUnlocked("arcade") {
		t.Fatal("a paid stage was unlocked on a fresh profile")
	}
}

func TestApplyRecordsStatsPointsAndProgress(t *testing.T) {
	m, store := newTestModel(t)

	m.apply(passResult("Home Row", 0, 130))

	stats := m.prof.Stats(typing.GameID)
	if stats.Plays != 1 {
		t.Errorf("Plays = %d, want 1", stats.Plays)
	}
	if stats.LessonsCleared != 1 {
		t.Errorf("LessonsCleared = %d, want 1", stats.LessonsCleared)
	}
	if stats.BestWPM != 25 {
		t.Errorf("BestWPM = %v, want 25", stats.BestWPM)
	}
	if stats.SecondsPlayed != 90 {
		t.Errorf("SecondsPlayed = %d, want 90", stats.SecondsPlayed)
	}
	if m.prof.PointsEarned != 130 {
		t.Errorf("PointsEarned = %d, want 130", m.prof.PointsEarned)
	}
	if m.err != nil {
		t.Errorf("apply recorded an error: %v", m.err)
	}

	// apply must persist, so a power cut straight after a lesson costs nothing.
	saved, err := store.Load("kiddo")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if saved.PointsEarned != 130 {
		t.Fatalf("saved PointsEarned = %d, want 130", saved.PointsEarned)
	}
}

func TestApplyOnlyAdvancesProgressForTheNextLesson(t *testing.T) {
	m, _ := newTestModel(t)

	m.apply(passResult("Home Row", 0, 100))
	// Replaying a lesson already cleared still pays, but does not skip ahead.
	m.apply(passResult("Home Row", 0, 100))

	if got := m.prof.Stats(typing.GameID).LessonsCleared; got != 1 {
		t.Fatalf("LessonsCleared = %d after replaying lesson 0, want 1", got)
	}
	if m.prof.PointsEarned != 200 {
		t.Fatalf("PointsEarned = %d, want 200", m.prof.PointsEarned)
	}

	// Skipping ahead is not possible either: only the next lesson advances it.
	m.apply(passResult("Real Words", 2, 100))
	if got := m.prof.Stats(typing.GameID).LessonsCleared; got != 1 {
		t.Fatalf("LessonsCleared = %d after passing lesson 2 out of order, want 1", got)
	}

	m.apply(passResult("Top and Bottom", 1, 100))
	if got := m.prof.Stats(typing.GameID).LessonsCleared; got != 2 {
		t.Fatalf("LessonsCleared = %d after passing lesson 1, want 2", got)
	}
}

func TestApplyAbandonedRunDoesNotAdvanceProgress(t *testing.T) {
	m, _ := newTestModel(t)

	r := passResult("Home Row", 0, 40)
	r.Completed = false
	m.apply(r)

	if got := m.prof.Stats(typing.GameID).LessonsCleared; got != 0 {
		t.Fatalf("LessonsCleared = %d after an abandoned run, want 0", got)
	}
	if m.prof.PointsEarned != 40 {
		t.Fatalf("PointsEarned = %d, want the 40 that were earned", m.prof.PointsEarned)
	}
}

func TestApplyGrantsAchievementsAndUnlocksStages(t *testing.T) {
	m, _ := newTestModel(t)

	m.apply(passResult("Home Row", 0, 300))

	if len(m.lastEarned) == 0 {
		t.Fatal("passing a first lesson granted no achievements")
	}
	if !m.prof.HasAchievement("first-lesson") {
		t.Error("first-lesson achievement not granted")
	}
	if !m.prof.HasAchievement("wpm-20") {
		t.Error("wpm-20 achievement not granted at 25 wpm")
	}
	if !m.prof.IsUnlocked("arcade") {
		t.Error("arcade stage not unlocked at 300 points")
	}

	var sawArcade bool
	for _, s := range m.lastUnlocks {
		if s.ID == "arcade" {
			sawArcade = true
		}
	}
	if !sawArcade {
		t.Error("the arcade unlock was not reported on the results screen")
	}

	// Badges are granted once. A second identical run reports nothing new.
	m.apply(passResult("Top and Bottom", 1, 10))
	for _, a := range m.lastEarned {
		if a.ID == "first-lesson" {
			t.Fatal("first-lesson granted twice")
		}
	}
}

// TestViewsRenderInEveryState drives the launcher the way a player would and
// renders each screen. It is a smoke test: View has no return value to assert
// against, but a nil map or a bad index in a format string panics here rather
// than on the laptop.
func TestViewsRenderInEveryState(t *testing.T) {
	m, _ := newTestModel(t)
	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if got := model.View(); !strings.Contains(got, "Typing Trainer") {
		t.Fatalf("menu does not list the typing game:\n%s", got)
	}

	// Into the game, and type the first line of the first lesson correctly.
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // open Typing Trainer
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // start lesson 0
	for _, r := range typing.Lessons[0].Lines[0] {
		if r == ' ' {
			model, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
			continue
		}
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	play := model.View()
	if !strings.Contains(play, "line 2 of") {
		t.Fatalf("typing a full line did not advance to line 2:\n%s", play)
	}

	// Bail out of the lesson: the launcher should show results.
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("leaving a lesson produced no command")
	}
	model, _ = model.Update(cmd())
	results := model.View()
	if !strings.Contains(results, "points") {
		t.Fatalf("results screen does not mention points:\n%s", results)
	}

	// Back to the menu, then into Progress.
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	progress := model.View()
	for _, want := range []string{"The Ladder", "Badges", "Steam"} {
		if !strings.Contains(progress, want) {
			t.Fatalf("progress screen is missing %q:\n%s", want, progress)
		}
	}
}

func TestMenuLocksGamesBehindTheirStage(t *testing.T) {
	m, _ := newTestModel(t)

	for _, item := range m.items {
		if item.kind == itemGame && item.game.ID() == typing.GameID && item.locked {
			t.Fatal("the typing game is locked on a fresh profile")
		}
	}
	// Every menu ends with the two housekeeping rows.
	if n := len(m.items); n < 3 {
		t.Fatalf("menu has %d items, want at least the game plus progress and quit", n)
	}
	if m.items[len(m.items)-1].kind != itemQuit {
		t.Error("quit is not the last menu item")
	}
}
