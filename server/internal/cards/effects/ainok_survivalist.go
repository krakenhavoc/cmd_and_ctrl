package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ainok Survivalist — Creature — Dog Shaman {1}{G}, 2/1 (Murders at
// Karlov Manor Commander):
//
//	"Megamorph {1}{G}
//	 When this creature is turned face up, destroy target artifact or
//	 enchantment an opponent controls."
//
// Megamorph is the keyword's (CR 702.37b); the trigger is Rambunctious
// Mutt's destroy clause on the turned-face-up event (CR 708.8). A card
// cast face down has no abilities (CR 708.2a), so only turning it up
// triggers.
//
// No simplification.
func init() {
	up := WhenThisIsTurnedFaceUp("Ainok Survivalist — destroy target artifact or enchantment an opponent controls",
		destroyFirstLegalTarget)
	up.Targets = TargetPermanent("target artifact or enchantment an opponent controls",
		Or(Artifact(), Enchantment()), OpponentControls())
	Register(Spec{
		OracleID:         "7a5dae5f-29ab-4996-b236-c754de8769bc",
		Name:             "Ainok Survivalist",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Megamorph("{1}{G}")},
		Triggered:        []game.TriggeredAbility{up},
	})
}
