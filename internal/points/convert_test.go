package points

import "testing"

func TestFromScore(t *testing.T) {
	cases := []struct {
		score     int
		completed bool
		want      int
	}{
		{0, true, 0},
		{-5, true, 0},
		{5, false, 0},  // an abandoned round with almost nothing done pays nothing
		{5, true, 1},   // but a pass always pays something
		{14, false, 0}, // rounds to nearest: just under half a point
		{15, false, 1}, // half a point rounds up
		{30, true, 1},
		{126, true, 4},
		{150, true, 5},
	}
	for _, c := range cases {
		if got := FromScore(c.score, c.completed); got != c.want {
			t.Errorf("FromScore(%d, %v) = %d, want %d", c.score, c.completed, got, c.want)
		}
	}
}
