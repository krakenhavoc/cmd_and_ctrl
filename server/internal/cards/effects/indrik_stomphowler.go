package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Indrik Stomphowler — Creature — Beast {4}{G}, 4/4:
//
//	"When this creature enters, destroy target artifact or enchantment."
//
// A mandatory enters trigger: with no artifact or enchantment anywhere it
// has no target and is removed from the stack (CR 603.3d); it may target
// the controller's own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de47a1e8-9c69-4af6-9d72-1bdd41352b32",
		Name:         "Indrik Stomphowler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Indrik Stomphowler — destroy target artifact or enchantment", destroyChosenPermanent),
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment()))),
		},
	})
}
