package game

// random_number.go — "choose a 10-digit number at random" (Duelists'
// Convocation International, #5948): a number drawn from the keyed pick
// stream (rng.go), so undo and a restore replay it exactly, and nobody
// chooses it.

// ChooseNumberAtRandomForEffect draws a `digits`-digit number at random,
// every such number equally likely: a first digit from 1 to 9 (a
// 10-digit number has no leading zero) and the rest from 0 to 9. It
// returns the digits, most significant first; nil for fewer than one
// digit. Nothing is emitted: the card says what it does with the number.
//
// Caller must hold g.mu (write).
func (g *Game) ChooseNumberAtRandomForEffect(d RandomDraw, digits int) []int {
	if digits <= 0 {
		return nil
	}
	r := g.randForLocked(rngStream{kind: rngStreamPick, player: d.Player, source: d.Source})
	out := make([]int, digits)
	for i := range out {
		if i == 0 && digits > 1 {
			out[i] = r.IntN(9) + 1
			continue
		}
		out[i] = r.IntN(10)
	}
	return out
}
