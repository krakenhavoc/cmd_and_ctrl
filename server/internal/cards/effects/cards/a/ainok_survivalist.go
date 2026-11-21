package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	up := effects.WhenThisIsTurnedFaceUp("Ainok Survivalist — destroy target artifact or enchantment an opponent controls",
		effects.DestroyFirstLegalTarget)
	up.Targets = effects.TargetPermanent("target artifact or enchantment an opponent controls",
		effects.Or(effects.Artifact(), effects.Enchantment()), effects.OpponentControls())
	effects.Register(effects.Spec{
		OracleID:         "7a5dae5f-29ab-4996-b236-c754de8769bc",
		Name:             "Ainok Survivalist",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{1}{G}")},
		Triggered:        []game.TriggeredAbility{up},
	})
}
