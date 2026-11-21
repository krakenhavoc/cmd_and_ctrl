package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "de47a1e8-9c69-4af6-9d72-1bdd41352b32",
		Name:         "Indrik Stomphowler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Indrik Stomphowler — destroy target artifact or enchantment", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))),
		},
	})
}
