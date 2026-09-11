package pseudocode

import (
	"strconv"
	"testing"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
)

// A program traced wrong goes to the back of the queue for one more go, so
// the correction screen -- which leaves the program up next to what it prints
// -- is something to act on rather than just something to read. Once only.

func TestAMistracedProgramComesRoundAgain(t *testing.T) {
	rd := newRound(Levels[0], games.NewTestRand(1))
	started := len(rd.programs)
	first := oneLine(rd.current())

	rd.submit("99999")
	if !rd.requeued {
		t.Fatal("a wrong trace was not sent round again")
	}
	if len(rd.programs) != started+1 {
		t.Fatalf("round has %d programs, want the missed one added back", len(rd.programs))
	}
	if got := oneLine(rd.programs[started]); got != first {
		t.Errorf("program sent round is %q, want %q", got, first)
	}

	for i := 1; i < started; i++ {
		if !rd.submit(strconv.Itoa(rd.current().Answer)) {
			t.Fatalf("program %d graded wrong", i)
		}
	}
	if rd.done {
		t.Fatal("round finished without asking the missed program again")
	}
	if got := oneLine(rd.current()); got != first {
		t.Fatalf("after the first pass the program is %q, want the missed one", got)
	}
	if !rd.submit(strconv.Itoa(rd.current().Answer)) {
		t.Fatal("the second attempt was graded wrong")
	}
	if !rd.done {
		t.Fatal("round did not finish after the second attempt")
	}
	// Right on the second go counts, and so does the fact that it took two.
	if got, want := rd.accuracy(), float64(started)/float64(started+1); got != want {
		t.Errorf("accuracy = %v, want %v", got, want)
	}
}

func TestAProgramIsOnlySentRoundOnce(t *testing.T) {
	rd := newRound(Levels[0], games.NewTestRand(2))
	started := len(rd.programs)

	for i := 0; i < started; i++ {
		rd.submit("99999")
	}
	if len(rd.programs) != 2*started {
		t.Fatalf("round has %d programs, want %d", len(rd.programs), 2*started)
	}
	for i := 0; i < started; i++ {
		rd.submit("99999")
		if rd.requeued {
			t.Fatal("a second attempt was sent round a third time")
		}
	}
	if !rd.done {
		t.Fatal("round did not finish once the second attempts were used up")
	}
	if len(rd.programs) != 2*started {
		t.Fatalf("round grew to %d programs after the second pass", len(rd.programs))
	}
	if rd.accuracy() != 0 {
		t.Errorf("accuracy = %v, want 0 when nothing was ever right", rd.accuracy())
	}
}

func TestNotesListAMissedProgramOnce(t *testing.T) {
	rd := newRound(Levels[0], games.NewTestRand(3))
	started := len(rd.programs)

	rd.submit("99999") // miss the first
	for i := 1; i < started; i++ {
		rd.submit(strconv.Itoa(rd.current().Answer))
	}
	rd.submit("99999") // and miss it again on its second go

	if notes := rd.notes(); len(notes) != 1 {
		t.Fatalf("notes = %v, want the one missed program listed once", notes)
	}
}

// A round can never run away: every program is asked at most twice, so the
// worst case is exactly double the length it started at.
func TestARoundCanAtWorstDoubleInLength(t *testing.T) {
	for _, l := range Levels {
		rd := newRound(l, games.NewTestRand(4))
		started := len(rd.programs)
		for !rd.done {
			rd.submit("-99999") // never right anywhere
		}
		if got := len(rd.programs); got != 2*started {
			t.Errorf("level %q: round of %d became %d, want %d",
				l.Title, started, got, 2*started)
		}
	}
}
