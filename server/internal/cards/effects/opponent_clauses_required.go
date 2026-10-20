package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_clauses_required.go — the REQUIRED sibling of
// opponent_clauses.go's "up to one" per-opponent clauses (#4128).

// ForNthOpponent binds a clause to the n-th seat (1-based) after the
// chooser's in turn order, REQUIRED while that seat holds a live
// opponent (game.TargetSpec.OpponentSlot). The candidate gate is the
// engine's, not a closure, so the legal set, the announce check and the
// CR 608.2b re-check all ask "does THAT player control it?".
func ForNthOpponent(n int, spec *game.TargetSpec) *game.TargetSpec {
	spec.OpponentSlot = n
	return spec
}

// perOpponentRequiredClauses is "for each opponent, <verb> target <noun>
// that player controls" (Blatant Thievery): one REQUIRED clause per
// possible opponent. A seat that is not at the table, or whose player
// has been eliminated, is not an opponent, so its clause is not required
// (see game/opponent_slot.go); a live opponent with nothing legal to
// target makes the spell uncastable (CR 601.2c).
func perOpponentRequiredClauses(noun string, what ...CardPredicate) *game.TargetSpec {
	var out []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		out = append(out, ForNthOpponent(n, TargetPermanent(
			"target "+noun+" controlled by the "+nthOpponentWords[n]+" opponent after you in turn order", what...)))
	}
	return Clauses(out[0], out[1:]...)
}
