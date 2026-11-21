package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nantuko Vigilante — Creature — Insect Druid Mutant {3}{G}, 3/2:
//
//	"Morph {1}{G}
//	 When this creature is turned face up, destroy target artifact or
//	 enchantment."
//
// Mandatory: the target is chosen as the trigger goes on the stack
// (CR 603.3d), and the trigger is dropped if nothing qualifies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "28ea78dc-7ad6-4006-aaf8-4b25057ec019",
		Name:             "Nantuko Vigilante",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisIsTurnedFaceUp("Nantuko Vigilante — destroy target artifact or enchantment", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))),
		},
	})
}
