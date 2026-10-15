package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Laccolith Whelp — Creature — Beast {R}, 1/1:
//
//	"Whenever this creature becomes blocked, you may have it deal damage equal to its power to
//	 target creature. If you do, this creature assigns no combat damage this turn."
//
// #3531: "becomes blocked" is the EventBecomesBlocked the block declaration's lock-in emits once per
// blocked attacker (CR 509.1h). The target is chosen as the trigger goes on the stack, "you may" is
// asked as it resolves, and the damage uses its power as it is then. On yes the creature's own combat
// damage is stopped by the CR 510.1a gate, so the blockers it was going to hit take none from it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "b16b10ec-a6a3-4bda-a561-9b7bb8eee194",
		Name:         "Laccolith Whelp",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			laccolithBlocked("Laccolith Whelp"),
		},
	})
}
