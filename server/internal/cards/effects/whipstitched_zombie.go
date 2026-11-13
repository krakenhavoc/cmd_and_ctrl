package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whipstitched Zombie — Creature — Zombie {1}{B}, 2/2 (Prophecy,
// #5617):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {B}."
//
// The Elder Dragons' upkeep (UpkeepSacrificeUnlessYouPay) at {B}: the
// controller is asked as the trigger resolves, and the table waits for
// the answer. Declining, or a "pay" they can't fund, sacrifices it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c498f0ff-812a-4106-834b-f43a0a7e0bac",
		Name:         "Whipstitched Zombie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("Whipstitched Zombie", "{B}"),
		},
	})
}
