package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fight to the Death — Instant {R}{W} (Marvel Universe, #4460):
//
//	"Destroy all blocking creatures and all blocked creatures."
//
// A creature is blocking from the moment it is declared as a blocker
// until combat is cleared (CR 506.4), and an attacker stays blocked after
// every creature blocking it has left combat (CR 509.1h); the two
// predicates are the ones Glyph of Destruction and Roar of Jukai read.
// Outside combat nothing matches. It is a destruction, so indestructible
// creatures live.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f885ffc-7245-4f00-a903-6e088007c635",
		Name:         "Fight to the Death",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Or(BlockingCreature(), BlockedCreature())}.Apply(ctx)
		},
	})
}
