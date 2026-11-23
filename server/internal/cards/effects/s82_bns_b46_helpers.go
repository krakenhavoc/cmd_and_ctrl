package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// opponentSpellsCastThisTurn is Delightful Discovery's "for each spell
// your opponents have cast this turn": the per-turn cast tally of every
// other seat still in the game. The spell being priced is not in its
// caster's tally yet, and the caster is excluded anyway.
func OpponentSpellsCastThisTurn(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	n := 0
	for _, p := range q.Game.Seats {
		if p == nil || p.ID == q.Controller || p.Eliminated {
			continue
		}
		n += q.Game.CastTallyFor(p.ID).Total
	}
	return n
}
