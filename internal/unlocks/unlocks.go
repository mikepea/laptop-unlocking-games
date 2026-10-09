// Package unlocks describes the ladder the laptop climbs as points are earned:
// it starts as a single game on a TTY and ends as a full desktop with Steam.
//
// Nothing here touches the operating system. A Stage is a declaration of intent
// that the launcher renders and that a future provisioning layer will act on.
package unlocks

// Stage is one rung of the ladder.
type Stage struct {
	ID    string
	Title string
	// Blurb is written for the player, not the parent.
	Blurb string
	// Cost is the lifetime points needed to reach this stage. A point is a
	// minute of screen time and takes about a minute of play to earn (see
	// points.ScorePerPoint), so a cost is also roughly minutes of play.
	Cost int
}

// Ladder is the ordered list of stages. Order matters: the launcher shows the
// first stage the player has not yet reached as "next up".
var Ladder = []Stage{
	{
		ID:    "typing",
		Title: "Typing Trainer",
		Blurb: "Where every laptop starts. Learn the keys.",
		Cost:  0,
	},
	{
		ID:    "arcade",
		Title: "The Arcade",
		Blurb: "Maths Sprint opens up. Number facts against the clock.",
		Cost:  8,
	},
	{
		ID:    "spelling",
		Title: "The Spelling Bee",
		Blurb: "A word appears, then vanishes. Write it down.",
		Cost:  18,
	},
	{
		ID:    "codebreaker",
		Title: "The Code Breaker",
		Blurb: "A hidden code, and enough clues to work it out.",
		Cost:  30,
	},
	{
		ID:    "pseudocode",
		Title: "Trace the Code",
		Blurb: "Programs to read. Work out what they print before they tell you.",
		Cost:  38,
	},
	{
		ID:    "files",
		Title: "Your Own Files",
		Blurb: "A place to keep things, and a way to look around it.",
		Cost:  47,
	},
	{
		ID:    "quest",
		Title: "Shell Quest",
		Blurb: "A pretend command line, with something hidden in it.",
		Cost:  67,
	},
	{
		ID:    "shell",
		Title: "The Shell",
		Blurb: "The real command line. Everything Shell Quest taught you, for keeps.",
		Cost:  100,
	},
	{
		ID:    "editor",
		Title: "The Editor",
		Blurb: "Write text. Write code. Break things and fix them.",
		Cost:  140,
	},
	{
		ID:    "network",
		Title: "The Wider World",
		Blurb: "A browser, and the internet behind it.",
		Cost:  187,
	},
	{
		ID:    "desktop",
		Title: "The Desktop",
		Blurb: "Windows, a mouse, a wallpaper of your choosing.",
		Cost:  250,
	},
	{
		ID:    "steam",
		Title: "Steam",
		Blurb: "Games, properly. The whole point, really.",
		Cost:  367,
	},
}

// ByID returns the stage with the given ID.
func ByID(id string) (Stage, bool) {
	for _, s := range Ladder {
		if s.ID == id {
			return s, true
		}
	}
	return Stage{}, false
}

// Reached returns the stages whose cost is covered by the given lifetime
// points, and the next stage still to come. reachedAll is true when the player
// has climbed the whole ladder, in which case next is the zero Stage.
func Reached(lifetimePoints int) (reached []Stage, next Stage, reachedAll bool) {
	for _, s := range Ladder {
		if lifetimePoints >= s.Cost {
			reached = append(reached, s)
			continue
		}
		return reached, s, false
	}
	return reached, Stage{}, true
}
