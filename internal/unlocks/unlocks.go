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
	// Cost is the lifetime points needed to reach this stage.
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
		Blurb: "More games appear in the menu.",
		Cost:  250,
	},
	{
		ID:    "files",
		Title: "Your Own Files",
		Blurb: "A place to keep things, and a way to look around it.",
		Cost:  600,
	},
	{
		ID:    "shell",
		Title: "The Shell",
		Blurb: "A real command line. This is how the machine talks back.",
		Cost:  1200,
	},
	{
		ID:    "editor",
		Title: "The Editor",
		Blurb: "Write text. Write code. Break things and fix them.",
		Cost:  2000,
	},
	{
		ID:    "network",
		Title: "The Wider World",
		Blurb: "A browser, and the internet behind it.",
		Cost:  3200,
	},
	{
		ID:    "desktop",
		Title: "The Desktop",
		Blurb: "Windows, a mouse, a wallpaper of your choosing.",
		Cost:  5000,
	},
	{
		ID:    "steam",
		Title: "Steam",
		Blurb: "Games, properly. The whole point, really.",
		Cost:  8000,
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
