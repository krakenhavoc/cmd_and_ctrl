package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4862: Register's gate for FromTargetPlayersGraveyard.
func TestFromTargetPlayersGraveyardRegistrationGate(t *testing.T) {
	cards := func() *game.TargetSpec {
		return TargetCardInGraveyard("up to three target cards from their graveyard").WithCount(0, 3)
	}
	// The printed shape registers.
	checkFlatClauses("ok", Clauses(TargetPlayer("target player"), FromTargetPlayersGraveyard(cards(), 0)))

	mustPanic(t, "not a player-only clause", func() {
		checkFlatClauses("any target", Clauses(TargetAny(), FromTargetPlayersGraveyard(cards(), 0)))
	})
	mustPanic(t, "only cards in graveyards", func() {
		checkFlatClauses("battlefield", Clauses(TargetPlayer("target player"),
			FromTargetPlayersGraveyard(TargetCreature("target creature").WithCount(0, 3), 0)))
	})
	mustPanic(t, "rule about its set of picks", func() {
		checkFlatClauses("single graveyard", Clauses(TargetPlayer("target player"),
			FromTargetPlayersGraveyard(cards().AllShare(FromASingleGraveyard()), 0)))
	})
	mustPanic(t, "one required earlier target", func() {
		checkFlatClauses("optional player", Clauses(TargetPlayer("target player").WithCount(0, 1),
			FromTargetPlayersGraveyard(cards(), 0)))
	})
	// The other relations still take at most one pick.
	mustPanic(t, "at most one pick", func() {
		checkFlatClauses("two welds", Clauses(TargetPermanent("target artifact", Artifact()),
			InGraveyardOfTargetsController(TargetCardInGraveyard("cards", Artifact()).WithCount(0, 2), 0)))
	})
}
