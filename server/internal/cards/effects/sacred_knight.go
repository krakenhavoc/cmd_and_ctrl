package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sacred Knight — Creature — Human Knight {3}{W}, 3/2 (Portal, #5630):
//
//	"This creature can't be blocked by black and/or red creatures."
//
// An evasion restriction, as Dauthi Horror's against white creatures:
// a creature that is black, red or both can't block it (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe8f5cc0-5849-437f-8d1f-cb191f2fc638",
		Name:         "Sacred Knight",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), Or(OfColor("B"), OfColor("R")), "black and/or red creatures"),
		},
	})
}
