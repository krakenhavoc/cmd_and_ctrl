package testst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4862: Register's gate for FromTargetPlayersGraveyard.
func TestFromTargetPlayersGraveyardRegistrationGate(t *testing.T) {
	cards := func() *game.TargetSpec {
		return effects.TargetCardInGraveyard("up to three target cards from their graveyard").WithCount(0, 3)
	}
	// The printed shape registers.
	effects.CheckFlatClauses("ok", effects.Clauses(effects.TargetPlayer("target player"), effects.FromTargetPlayersGraveyard(cards(), 0)))

	effectstest.MustPanic(t, "not a player-only clause", func() {
		effects.CheckFlatClauses("any target", effects.Clauses(effects.TargetAny(), effects.FromTargetPlayersGraveyard(cards(), 0)))
	})
	effectstest.MustPanic(t, "only cards in graveyards", func() {
		effects.CheckFlatClauses("battlefield", effects.Clauses(effects.TargetPlayer("target player"),
			effects.FromTargetPlayersGraveyard(effects.TargetCreature("target creature").WithCount(0, 3), 0)))
	})
	effectstest.MustPanic(t, "rule about its set of picks", func() {
		effects.CheckFlatClauses("single graveyard", effects.Clauses(effects.TargetPlayer("target player"),
			effects.FromTargetPlayersGraveyard(cards().AllShare(effects.FromASingleGraveyard()), 0)))
	})
	effectstest.MustPanic(t, "one required earlier target", func() {
		effects.CheckFlatClauses("optional player", effects.Clauses(effects.TargetPlayer("target player").WithCount(0, 1),
			effects.FromTargetPlayersGraveyard(cards(), 0)))
	})
	// The other relations still take at most one pick.
	effectstest.MustPanic(t, "at most one pick", func() {
		effects.CheckFlatClauses("two welds", effects.Clauses(effects.TargetPermanent("target artifact", effects.Artifact()),
			effects.InGraveyardOfTargetsController(effects.TargetCardInGraveyard("cards", effects.Artifact()).WithCount(0, 2), 0)))
	})
}
