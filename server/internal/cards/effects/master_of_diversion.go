package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master of Diversion — Creature — Human Scout {2}{W}, 2/2 (Magic 2014, #5563):
//
//	"Whenever this creature attacks, tap target creature defending player controls."
//
// Distracting Geist's attack trigger: the target is a creature the defending
// player controls, chosen as the trigger goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3402aa73-5e63-4158-b114-6f1d42ec8992",
		Name:         "Master of Diversion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			distractingGeistTap("Master of Diversion — tap target creature defending player controls"),
		},
	})
}
