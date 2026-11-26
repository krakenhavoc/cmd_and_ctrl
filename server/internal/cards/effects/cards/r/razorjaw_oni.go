package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Razorjaw Oni — Creature — Demon Spirit {3}{B}, 4/5 (Commander 2011, #5574):
//
//	"Black creatures can't block."
//
// One block rule (CR 509.1b), Bedlam's shape narrowed to black blockers: no
// black creature on the battlefield, whoever controls it, can be declared as a
// blocker while the Oni is on the battlefield, itself included. Colour is read
// from effective characteristics when the block is declared, so a creature
// that becomes black is covered and one that stops being black is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed8bc0f2-5a65-4eb4-bf88-3642a55b6889",
		Name:         "Razorjaw Oni",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnMatching(effects.And(effects.Creature(), effects.OfColor("B"))), effects.OnMatching(effects.Creature()), "Razorjaw Oni: black creatures can't block"),
		},
	})
}
