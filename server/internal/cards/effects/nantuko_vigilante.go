package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "28ea78dc-7ad6-4006-aaf8-4b25057ec019",
		Name:             "Nantuko Vigilante",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisIsTurnedFaceUp("Nantuko Vigilante — destroy target artifact or enchantment", destroyFirstLegalCardTarget),
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment()))),
		},
	})
}
