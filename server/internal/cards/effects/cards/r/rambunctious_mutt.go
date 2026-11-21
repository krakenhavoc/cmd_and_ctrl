package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rambunctious Mutt — Creature — Dog {3}{W}{W}, 3/4:
//
//	"When this creature enters, destroy target artifact or enchantment
//	 an opponent controls."
//
// Not a "may": with a legal target the trigger must take one, and with
// none it is removed (CR 603.3d). The target is re-checked as it
// resolves (CR 608.2b): one that changed control to you in response is
// no longer legal, and an indestructible one survives.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Rambunctious Mutt — destroy target artifact or enchantment an opponent controls",
		effects.DestroyFirstLegalTarget)
	enters.Targets = effects.TargetPermanent("target artifact or enchantment an opponent controls",
		effects.Or(effects.Artifact(), effects.Enchantment()), effects.OpponentControls())
	effects.Register(effects.Spec{
		OracleID:     "48a53d6e-0320-4b9c-b692-6d0fc00e46cc",
		Name:         "Rambunctious Mutt",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
