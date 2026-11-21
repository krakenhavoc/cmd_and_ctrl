package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Odric's Outrider — Creature — Human Knight {3}{W}, 2/4 (Innistrad: Midnight Hunt, S82):
//
//	"Whenever this creature or another creature you control dies, put a +1/+1 counter on target creature you control."
//
// Butcher of Malakir's "this or another" dies trigger. The target is chosen as the trigger goes on the stack, so the Outrider's own death can still pick another creature; with none it is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5a9a314-81a0-4b11-9cf9-daeaea98b33c",
		Name:         "Odric's Outrider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverACreatureYouControlDies("Odric's Outrider — put a +1/+1 counter on target creature you control",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
