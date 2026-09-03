package unlocks

import "testing"

func TestLadderIsOrderedByCost(t *testing.T) {
	// Reached walks the ladder and stops at the first stage it cannot afford,
	// so an out-of-order cost would silently hide every stage after it.
	for i := 1; i < len(Ladder); i++ {
		if Ladder[i].Cost < Ladder[i-1].Cost {
			t.Fatalf("stage %q costs %d, less than %q at %d",
				Ladder[i].ID, Ladder[i].Cost, Ladder[i-1].ID, Ladder[i-1].Cost)
		}
	}
}

func TestLadderIDsAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range Ladder {
		if seen[s.ID] {
			t.Errorf("duplicate stage ID %q", s.ID)
		}
		seen[s.ID] = true
	}
}

func TestReached(t *testing.T) {
	reached, next, all := Reached(0)
	if len(reached) != 1 || reached[0].ID != "typing" {
		t.Fatalf("at 0 points, reached = %v, want just typing", ids(reached))
	}
	if all {
		t.Fatal("at 0 points the ladder reported complete")
	}
	if next.ID != "arcade" {
		t.Fatalf("next = %q, want arcade", next.ID)
	}

	// Exactly the cost of a stage is enough to reach it. Asserted against the
	// ladder's own order rather than a hardcoded name, so inserting a rung is
	// not a test failure.
	arcade, _ := ByID("arcade")
	reached, next, _ = Reached(arcade.Cost)
	if len(reached) != 2 {
		t.Fatalf("at %d points, reached = %v, want typing and arcade", arcade.Cost, ids(reached))
	}
	if want := Ladder[2].ID; next.ID != want {
		t.Fatalf("next = %q, want %q", next.ID, want)
	}

	// Past the top of the ladder.
	top := Ladder[len(Ladder)-1]
	reached, next, all = Reached(top.Cost + 1)
	if !all {
		t.Fatal("past the top cost, the ladder did not report complete")
	}
	if len(reached) != len(Ladder) {
		t.Fatalf("reached %d stages, want all %d", len(reached), len(Ladder))
	}
	if next != (Stage{}) {
		t.Fatalf("next = %+v, want the zero Stage", next)
	}
}

func TestByID(t *testing.T) {
	if s, ok := ByID("shell"); !ok || s.Title == "" {
		t.Fatalf("ByID(shell) = (%+v, %v)", s, ok)
	}
	if _, ok := ByID("no-such-stage"); ok {
		t.Fatal("ByID found a stage that does not exist")
	}
}

func ids(stages []Stage) []string {
	out := make([]string, len(stages))
	for i, s := range stages {
		out[i] = s.ID
	}
	return out
}
