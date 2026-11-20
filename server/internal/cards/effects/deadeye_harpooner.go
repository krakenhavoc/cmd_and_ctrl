package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	ability := WhenThisEntersIfRevolt("Deadeye Harpooner — destroy target tapped creature an opponent controls",
		destroyFirstLegalCardTarget)
	ability.Targets = TargetCreature("target tapped creature an opponent controls", hobB3TappedCreature(), OpponentControls())
	Register(Spec{
		OracleID:     "959e1f72-4ba2-4c87-95d1-ca0265dc128c",
		Name:         "Deadeye Harpooner",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{ability},
	})
}
