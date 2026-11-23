package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadeye Harpooner — Creature — Dwarf Warrior {2}{W}, 2/2 (Aether Revolt,
// #5536):
//
//	"Revolt — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, destroy target tapped
//	 creature an opponent controls."
//
// Renegade Rallier's targeted revolt trigger with Harbinger of the
// Tides' "tapped creature an opponent controls" clause. Without revolt it
// does not trigger; a target untapped in response is illegal and the
// trigger does nothing (CR 608.2b).
//
// No simplification.
func init() {
	ability := effects.WhenThisEntersIfRevolt("Deadeye Harpooner — destroy target tapped creature an opponent controls",
		effects.DestroyFirstLegalCardTarget)
	ability.Targets = effects.TargetCreature("target tapped creature an opponent controls", effects.HobB3TappedCreature(), effects.OpponentControls())
	effects.Register(effects.Spec{
		OracleID:     "959e1f72-4ba2-4c87-95d1-ca0265dc128c",
		Name:         "Deadeye Harpooner",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{ability},
	})
}
