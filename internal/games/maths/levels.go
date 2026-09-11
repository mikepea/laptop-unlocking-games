package maths

import "math/rand/v2"

// Question is one thing to answer.
type Question struct {
	// Context is optional working shown above the question: the equation being
	// solved, the expression being collected up. Empty for plain arithmetic,
	// where the prompt says everything.
	Context string
	// Prompt is shown without the answer, e.g. "7 x 8" or "x". A prompt ending
	// in "?" is a sentence and is rendered as one; anything else is treated as
	// the left-hand side of an equals sign.
	Prompt string
	Answer Answer
}

// Level is a set of questions of one kind. Levels are cleared in order.
type Level struct {
	Title string
	Hint  string
	// Questions is how many are asked in a round.
	Questions int
	// MinAccuracy is the share that must be right to pass.
	MinAccuracy float64
	// Gen makes one question. It is called once per question in a round.
	Gen func(r *rand.Rand) Question
	// GenSteps makes one multi-step problem, as the sequence of questions that
	// walks through it. Levels set either this or Gen, never both. A round
	// never splits a problem across its end, so a level using this can run a
	// question or two past Questions rather than leave working half-done.
	GenSteps func(r *rand.Rand) []Question
}

// Chapter is a group of levels, and maps onto a chapter of the Art of Problem
// Solving Prealgebra book so that "do chapter 4 tonight" means the same thing
// at the kitchen table and in the game.
type Chapter struct {
	// Num is the chapter number in the book. Zero is the warm-up, which is not
	// in the book: it is the number facts everything else stands on.
	Num int
	// Title matches the book's chapter title, so the contents page is the
	// menu.
	Title string
	// Blurb is written for the player, not the parent.
	Blurb  string
	Levels []Level
}

// Chapters is the curriculum. The order is the book's order, which is also the
// order levels unlock in: a level opens when the one before it is passed, and
// that runs straight through a chapter boundary.
//
// Only the first nine chapters are here. They are the ones that are pure
// number and so fit in a single answer field; the book's geometry, statistics
// and counting chapters want figures drawn and are deliberately not built.
var Chapters = []Chapter{
	{
		Num:    0,
		Title:  "Warm-Up",
		Blurb:  "Number facts. Everything else is built on these.",
		Levels: warmUpLevels,
	},
	{
		Num:    1,
		Title:  "Properties of Arithmetic",
		Blurb:  "Rearranging sums to make them easy, and going below zero.",
		Levels: propertiesLevels,
	},
	{
		Num:    2,
		Title:  "Exponents",
		Blurb:  "Powers, and the shortcuts for multiplying and dividing them.",
		Levels: exponentLevels,
	},
	{
		Num:    3,
		Title:  "Number Theory",
		Blurb:  "What divides into what. Primes, factors and multiples.",
		Levels: numberTheoryLevels,
	},
	{
		Num:    4,
		Title:  "Fractions",
		Blurb:  "Parts of a whole, and how to do arithmetic with them.",
		Levels: fractionLevels,
	},
	{
		Num:    5,
		Title:  "Equations and Inequalities",
		Blurb:  "Letters standing in for numbers, and how to find them.",
		Levels: equationLevels,
	},
	{
		Num:    6,
		Title:  "Decimals",
		Blurb:  "Fractions written another way, with a point in them.",
		Levels: decimalLevels,
	},
	{
		Num:    7,
		Title:  "Ratios, Conversions and Rates",
		Blurb:  "Comparing amounts, changing units, and how fast things go.",
		Levels: ratioLevels,
	},
	{
		Num:    8,
		Title:  "Percents",
		Blurb:  "Hundredths, and what they do to prices.",
		Levels: percentLevels,
	},
	{
		Num:    9,
		Title:  "Square Roots",
		Blurb:  "Undoing a square, and tidying up what is left.",
		Levels: rootLevels,
	},
}

// Levels is every level in every chapter, flattened into the one ordered list
// the profile counts against. GameStats.LessonsCleared is an index into this,
// so a level may be appended to the end of a chapter but never inserted in
// front of one that is already there -- that would silently hand out, or take
// away, progress nobody earned.
var Levels = flattenChapters()

// chapterStart[i] is the index in Levels of the first level of Chapters[i]. It
// is what turns a position in the two-level menu into the flat index the
// profile and the launcher's Result talk in.
var chapterStart = startsOf(Chapters)

func flattenChapters() []Level {
	var all []Level
	for _, c := range Chapters {
		all = append(all, c.Levels...)
	}
	return all
}

func startsOf(chapters []Chapter) []int {
	starts := make([]int, len(chapters))
	n := 0
	for i, c := range chapters {
		starts[i] = n
		n += len(c.Levels)
	}
	return starts
}

// LevelIndex is the position in Levels of level j of chapter i.
func LevelIndex(chapter, level int) int { return chapterStart[chapter] + level }

// ChapterUnlocked reports whether any of a chapter's levels can be played.
// A chapter opens as soon as its first level does.
func ChapterUnlocked(chapter, cleared int) bool {
	return LevelUnlocked(chapterStart[chapter], cleared)
}

// ChapterCleared is how many of a chapter's levels have been passed.
func ChapterCleared(chapter, cleared int) int {
	done := cleared - chapterStart[chapter]
	return min(max(done, 0), len(Chapters[chapter].Levels))
}

// LevelUnlocked reports whether level i is open given how many are cleared.
func LevelUnlocked(i, cleared int) bool { return i <= cleared }
