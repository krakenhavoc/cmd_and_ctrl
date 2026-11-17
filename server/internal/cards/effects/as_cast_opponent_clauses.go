package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// as_cast_opponent_clauses.go — #5860: "for each opponent who has N or
// more poison counters as you cast this spell, … up to one target …
// from that player's graveyard" (Geth's Summons). The gate is the
// engine's (game/as_cast_opponent.go): live while the spell is being
// cast, fixed on the item once it is, so the CR 608.2b re-check never
// re-judges the poison.

// ForNthOpponentPoisonedAsCast binds a card clause to the n-th seat
// (1-based) after the chooser's in turn order: the candidate must be
// owned by that opponent, who must have had `minPoison` or more poison
// counters as the spell was cast (game.TargetSpec.AsCastOpponentSlot).
func ForNthOpponentPoisonedAsCast(n, minPoison int, spec *game.TargetSpec) *game.TargetSpec {
	spec.AsCastOpponentSlot = n
	spec.AsCastMinPoison = minPoison
	return spec
}

// perPoisonedOpponentGraveyardClauses is one "up to one target <noun>
// from that player's graveyard" clause per possible opponent, each live
// only for an opponent with `minPoison` or more poison counters as the
// spell was cast.
func perPoisonedOpponentGraveyardClauses(minPoison int, noun string, what ...CardPredicate) []*game.TargetSpec {
	var out []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		out = append(out, ForNthOpponentPoisonedAsCast(n, minPoison, TargetCardInGraveyard(
			"up to one target "+noun+" from the graveyard of the "+nthOpponentWords[n]+
				" opponent after you in turn order, if they had "+strconv.Itoa(minPoison)+" or more poison counters as you cast this spell", what...).WithCount(0, 1)))
	}
	return out
}
