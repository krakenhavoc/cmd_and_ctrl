package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rottenheart Ghoul — {3}{B} Creature — Zombie 2/4 (Shadows over Innistrad,
// S82 OLD b293):
//
//	"When this creature dies, target player discards a card."
//
// Skull Fracture's discard, aimed as the trigger goes on the stack (CR
// 603.3d); the target player chooses what to discard (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9535aa3-0c07-4fad-8dc9-9838a9ebd304",
		Name:         "Rottenheart Ghoul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Rottenheart Ghoul — target player discards a card", effects.TargetPlayerDiscards(1)),
				effects.TargetPlayer("target player")),
		},
	})
}
