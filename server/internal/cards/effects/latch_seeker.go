package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Latch Seeker — {1}{U}{U} Creature — Spirit 3/1 (Avacyn Restored, #5570):
//
//	"This creature can't be blocked."
//
// An evasion static on itself (CR 509.1b), Triton Shorestalker's row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c1f62a8-d905-48aa-aba1-4bcb522a7ef6",
		Name:         "Latch Seeker",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
