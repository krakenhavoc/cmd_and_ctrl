package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barrenton Cragtreads — Creature — Kithkin Scout {2}{W/U}{W/U}, 3/3
// (Shadowmoor, #5591):
//
//	"This creature can't be blocked by red creatures."
//
// A block rule, Shield Mare's: a blocker of that colour, as it is now, is
// refused (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "969233f8-a644-4297-a2bf-b1ca7a58bb0b",
		Name:         "Barrenton Cragtreads",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), And(Creature(), OfColor("R")), "red creatures"),
		},
	})
}
