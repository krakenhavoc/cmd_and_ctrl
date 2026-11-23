package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// acr_5466.go — Assassin's Creed batch 1 (#5466) helpers shared across test packages.

// Acrb1ConnectThenPostcombat attacks seat 1 with a creature of
// `typeLine` under seat 0 and walks on to the postcombat main phase.
func Acrb1ConnectThenPostcombat(t *testing.T, g *game.Game, typeLine string) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	dealer := PushDiesCreatureForTest(g, me.ID, "Dealer", "", typeLine, 2, 2)
	AttackWith(t, g, opp.ID, dealer)
	PassPriorityAroundTable(t, g)
	AdvanceTo(t, g, game.StepPostcombatMain)
}
