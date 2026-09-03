// Package achievements defines the badges a player can earn and the rules that
// grant them.
//
// Rules are evaluated after a result has been folded into the profile, so a
// rule can look at either the run that just happened or the running totals.
package achievements

import (
	"github.com/mikepea/laptop-unlocking-games/internal/games"
	"github.com/mikepea/laptop-unlocking-games/internal/games/typing"
	"github.com/mikepea/laptop-unlocking-games/internal/profile"
)

// Achievement is one badge.
type Achievement struct {
	ID          string
	Title       string
	Description string
	// Check reports whether the badge is earned. It is only consulted for
	// badges the player does not already hold.
	Check func(p *profile.Profile, r games.Result) bool
}

// All is every achievement in this build, in the order they are listed to the
// player.
var All = []Achievement{
	{
		ID:          "first-lesson",
		Title:       "First Steps",
		Description: "Pass your first typing lesson.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return r.Completed
		},
	},
	{
		ID:          "flawless",
		Title:       "Flawless",
		Description: "Finish a lesson without a single mistake.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return r.Completed && r.Accuracy >= 1
		},
	},
	{
		ID:          "wpm-20",
		Title:       "Getting Going",
		Description: "Pass a lesson at 20 words per minute.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return r.Completed && r.WPM >= 20
		},
	},
	{
		ID:          "wpm-35",
		Title:       "Quick Fingers",
		Description: "Pass a lesson at 35 words per minute.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return r.Completed && r.WPM >= 35
		},
	},
	{
		ID:          "wpm-50",
		Title:       "Touch Typist",
		Description: "Pass a lesson at 50 words per minute.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return r.Completed && r.WPM >= 50
		},
	},
	{
		ID:          "persistent",
		Title:       "Persistent",
		Description: "Play ten times.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return totalPlays(p) >= 10
		},
	},
	{
		ID:          "stayer",
		Title:       "The Long Haul",
		Description: "Spend an hour at the keyboard.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return totalSeconds(p) >= 3600
		},
	},
	{
		ID:          "punctual",
		Title:       "Full Stop",
		Description: "Pass the punctuation lesson.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return r.Completed && r.GameID == typing.GameID && r.Round == "Punctuation"
		},
	},
	{
		ID:          "graduate",
		Title:       "Graduate",
		Description: "Pass every typing lesson.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return p.Stats(typing.GameID).LessonsCleared >= len(typing.Lessons)
		},
	},
	{
		ID:          "thousandaire",
		Title:       "Thousandaire",
		Description: "Earn a thousand points.",
		Check: func(p *profile.Profile, r games.Result) bool {
			return p.PointsEarned >= 1000
		},
	},
}

// ByID looks up an achievement definition.
func ByID(id string) (Achievement, bool) {
	for _, a := range All {
		if a.ID == id {
			return a, true
		}
	}
	return Achievement{}, false
}

// Evaluate grants any newly earned achievements and returns them. The profile
// must already reflect the result being evaluated.
func Evaluate(p *profile.Profile, r games.Result) []Achievement {
	var earned []Achievement
	for _, a := range All {
		if p.HasAchievement(a.ID) {
			continue
		}
		if a.Check != nil && a.Check(p, r) && p.GrantAchievement(a.ID) {
			earned = append(earned, a)
		}
	}
	return earned
}

func totalPlays(p *profile.Profile) int {
	n := 0
	for _, s := range p.Games {
		n += s.Plays
	}
	return n
}

func totalSeconds(p *profile.Profile) int {
	n := 0
	for _, s := range p.Games {
		n += s.SecondsPlayed
	}
	return n
}
