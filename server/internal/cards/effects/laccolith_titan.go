package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Laccolith Titan — Creature — Beast {5}{R}{R}, 6/6:
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
		OracleID:     "0d9ac5e0-2b4e-4e7b-937e-9e9bfce7e96b",
		Name:         "Laccolith Titan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			laccolithBlocked("Laccolith Titan"),
		},
	})
}
