package points

// ScorePerPoint is how much in-game score buys one point. A point is a minute
// of screen time, and a well-played round of any game scores roughly 100 in
// three or four minutes, so at 30 a round pays three or four points and an
// hour's worth of points takes about an hour of play. This is the one number
// to change to make points easier or harder to earn; the unlocks ladder is
// priced in points, so it moves with it.
const ScorePerPoint = 30

// FromScore turns a game's score into the points it pays, rounding to the
// nearest point. A round that was passed always pays at least one, so a pass
// is never worth nothing; an abandoned round with a small score can.
func FromScore(score int, completed bool) int {
	if score <= 0 {
		return 0
	}
	n := (score + ScorePerPoint/2) / ScorePerPoint
	if n == 0 && completed {
		n = 1
	}
	return n
}
