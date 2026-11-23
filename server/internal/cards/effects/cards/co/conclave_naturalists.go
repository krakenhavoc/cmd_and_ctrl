package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conclave Naturalists — Creature — Dryad {4}{G}, 4/4 (Magic Origins, #5549):
//
//	"When this creature enters, you may destroy target artifact or enchantment."
//
// Reclamation Sage's entry: the target is chosen as the trigger goes on
// the stack, and the "you may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb5de977-d420-41cc-85f0-6ffee8bbeff1",
		Name:         "Conclave Naturalists",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Conclave Naturalists — destroy target artifact or enchantment", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))),
				"Conclave Naturalists — destroy an artifact or enchantment?"),
		},
	})
}
