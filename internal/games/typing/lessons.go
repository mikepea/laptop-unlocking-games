package typing

// Lesson is one set of lines to type. Lessons are cleared in order: a lesson
// opens once the one before it has been passed.
type Lesson struct {
	Title string
	// Hint is one line of coaching shown above the text.
	Hint string
	// MinAccuracy is the share of keystrokes that must be right to pass.
	MinAccuracy float64
	Lines       []string
}

// Lessons is the ordered curriculum. Lines are kept under 48 characters so they
// fit comfortably on an 80-column TTY with room for the frame.
var Lessons = []Lesson{
	{
		Title:       "Home Row",
		Hint:        "Fingers on asdf jkl; don't look down.",
		MinAccuracy: 0.90,
		Lines: []string{
			"asdf jkl; asdf jkl; asdf jkl;",
			"a sad lad; a fall; ask dad;",
			"all lads fall; dad asks;",
			"salad flask; a glad lass;",
			"half a flask; a small salad;",
		},
	},
	{
		Title:       "Top and Bottom",
		Hint:        "Reach up and down, then come straight back to home row.",
		MinAccuracy: 0.90,
		Lines: []string{
			"quiet river towns wept",
			"the zebra can box every",
			"pack my box with jugs",
			"vex quick brown foxes",
			"why did the crow zoom",
			"jump over the lazy dog",
		},
	},
	{
		Title:       "Real Words",
		Hint:        "Words you already know. Aim for smooth, not fast.",
		MinAccuracy: 0.92,
		Lines: []string{
			"the quick brown fox jumps over it",
			"a laptop that unlocks when you learn",
			"every game has a way to win",
			"keep your eyes on the screen",
			"slow is smooth and smooth is fast",
			"one more line and you are done",
		},
	},
	{
		Title:       "Capitals",
		Hint:        "Hold Shift with the opposite hand from the letter.",
		MinAccuracy: 0.92,
		Lines: []string{
			"Monday Tuesday Wednesday Thursday",
			"London Paris Tokyo Cairo Lima",
			"The Moon Is Not Made Of Cheese",
			"Ada Lovelace Wrote The First Program",
			"Linux Was Written By Many People",
			"Press Shift And The Letter Together",
		},
	},
	{
		Title:       "Punctuation",
		Hint:        "Commas, full stops and apostrophes count too.",
		MinAccuracy: 0.92,
		Lines: []string{
			"Hello, world. That's the first one.",
			"Wait - what's that noise?",
			"It's fine; nothing is on fire.",
			"\"Stop,\" she said, \"and look.\"",
			"Don't panic. Read the error message.",
			"Three things: keys, screen, patience.",
		},
	},
	{
		Title:       "Numbers and Symbols",
		Hint:        "The top row. Shift makes the symbols.",
		MinAccuracy: 0.90,
		Lines: []string{
			"1 2 3 4 5 6 7 8 9 0",
			"12 + 30 = 42 and 100 - 58 = 42",
			"50% of 84 is 42 (always 42)",
			"user@example.com #1 $5 &more",
			"if x > 10 { print(x * 2) }",
			"2026-09-02 07:30 (90 minutes)",
		},
	},
	{
		Title:       "Real Sentences",
		Hint:        "Everything at once. This is what typing actually is.",
		MinAccuracy: 0.93,
		Lines: []string{
			"A computer does exactly what you tell it.",
			"That is the problem, and also the point.",
			"Save your work before you try something.",
			"Every expert was once completely stuck.",
			"Read the message; it usually says why.",
			"Then fix one thing and run it again.",
		},
	},
}

// LessonUnlocked reports whether lesson index i is open given how many lessons
// have been cleared. The first lesson is always open.
func LessonUnlocked(i, cleared int) bool { return i <= cleared }
