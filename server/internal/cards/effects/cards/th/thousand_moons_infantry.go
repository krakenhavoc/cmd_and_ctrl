package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thousand Moons Infantry — Creature — Human Soldier {2}{W}, 2/4 (The
// Lost Caverns of Ixalan, #4470):
//
//	"Untap this creature during each other player's untap step."
//
// Bender's Waterskin's clause on a creature: a permission on the untap
// step's turn-based action (CR 502.3), not a triggered ability, so it
// uses no stack and nothing can respond (untap_step.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f064a39e-ebb3-47b5-bd06-f39b02b95eaf",
		Name:         "Thousand Moons Infantry",
		Completeness: effects.CompletenessFull,
		UntapStep: []game.UntapStepPermission{
			effects.UntapSelfDuringEachOtherPlayersUntapStep(
				"Thousand Moons Infantry — untap this creature"),
		},
	})
}
