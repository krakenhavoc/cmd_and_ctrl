package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bala Ged Scorpion — Creature — Scorpion {3}{B}, 2/3 (Rise of the Eldrazi,
// S82 OLD b485):
//
//	"When this creature enters, you may destroy target creature with power 1
//	 or less."
//
// An optional enters trigger (CR 603.5), Fleshpulper Giant's shape: the target
// is chosen as the trigger goes on the stack, with power read through the layers
// at announcement and again at resolution (CR 608.2b), and the "you may" is asked
// as it resolves. The Scorpion's own power is 2, so it is no target for itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "753b4258-bc1d-47f4-b994-356507b0816a",
		Name:         "Bala Ged Scorpion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Bala Ged Scorpion — you may destroy target creature with power 1 or less", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target creature with power 1 or less", effects.PowerLE(1))),
				"Bala Ged Scorpion — destroy the creature?"),
		},
	})
}
