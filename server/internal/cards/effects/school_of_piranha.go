package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// School of Piranha — Creature — Fish {1}{U}, 3/3 (Exodus, #5626):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {1}{U}."
//
// Pit Raptor's upkeep through the shared UpkeepSacrificeUnlessYouPay:
// the controller is asked as the trigger resolves, and declining, or a
// "yes" they cannot fund, sacrifices it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fad69c20-d6f4-4cf5-8e4c-bbda8d268edd",
		Name:         "School of Piranha",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("School of Piranha", "{1}{U}"),
		},
	})
}
