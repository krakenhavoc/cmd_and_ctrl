package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigiled Skink — Creature — Lizard {1}{R}, 2/1 (Journey into Nyx, #5558):
//
//	"Whenever this creature attacks, scry 1. (Look at the top card of
//	 your library. You may put that card on the bottom.)"
//
// Lothlórien Lookout's attack scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c17f6c3-48d3-4bd0-b0dd-76229d9c5b09",
		Name:         "Sigiled Skink",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Sigiled Skink — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
