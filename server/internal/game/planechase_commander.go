package game

// planechase_commander.go is the engine half of The Command Zone (ADR 0216
// slice 4, #4870).

// ResetCommanderCastsForEffect is "the number of times each commander has
// been cast from the command zone becomes 0": every player's commander tax
// tally (Player.CommanderCasts, CR 903.8) goes back to zero, so each
// commander's next cast from the command zone costs no extra.
//
// Caller must hold g.mu (write).
func (g *Game) ResetCommanderCastsForEffect() {
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for id := range p.CommanderCasts {
			delete(p.CommanderCasts, id)
		}
	}
}
