package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conclave Naturalists — Creature — Dryad {4}{G}, 4/4 (Magic Origins, #5549):
//
//	"When this creature enters, you may destroy target artifact or enchantment."
//
// Reclamation Sage's entry: the target is chosen as the trigger goes on
// the stack, and the "you may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb5de977-d420-41cc-85f0-6ffee8bbeff1",
		Name:         "Conclave Naturalists",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Conclave Naturalists — destroy target artifact or enchantment", destroyChosenPermanent),
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment()))),
				"Conclave Naturalists — destroy an artifact or enchantment?"),
		},
	})
}
