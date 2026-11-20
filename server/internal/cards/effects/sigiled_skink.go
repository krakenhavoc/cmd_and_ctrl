package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigiled Skink — Creature — Lizard {1}{R}, 2/1 (Journey into Nyx, #5558):
//
//	"Whenever this creature attacks, scry 1. (Look at the top card of
//	 your library. You may put that card on the bottom.)"
//
// Lothlórien Lookout's attack scry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c17f6c3-48d3-4bd0-b0dd-76229d9c5b09",
		Name:         "Sigiled Skink",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Sigiled Skink — scry 1", Do(Scry{N: 1})),
		},
	})
}
