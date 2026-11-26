package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marshdrinker Giant — Creature — Giant Warrior {3}{G}{G} 4/3:
//
//	"When this creature enters, destroy target Island or Swamp an opponent
//	 controls."
//
// A targeted enters trigger: the target is an Island or a Swamp (a land with
// either subtype, so a dual land with both qualifies) an opponent controls,
// chosen as the trigger is put on the stack (CR 603.3d); with none it is
// removed from the stack.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be109ed5-9e3e-40b3-9bf0-7caf7cdbb958",
		Name:         "Marshdrinker Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Marshdrinker Giant — destroy target Island or Swamp an opponent controls", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target Island or Swamp an opponent controls",
					effects.Land(), effects.OpponentControls(),
					effects.Or(effects.OfSubtype("Island"), effects.OfSubtype("Swamp")))),
		},
	})
}
