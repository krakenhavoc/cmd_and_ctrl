package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	enters := WhenThisEnters("Rambunctious Mutt — destroy target artifact or enchantment an opponent controls",
		destroyFirstLegalTarget)
	enters.Targets = TargetPermanent("target artifact or enchantment an opponent controls",
		Or(Artifact(), Enchantment()), OpponentControls())
	Register(Spec{
		OracleID:     "48a53d6e-0320-4b9c-b692-6d0fc00e46cc",
		Name:         "Rambunctious Mutt",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
