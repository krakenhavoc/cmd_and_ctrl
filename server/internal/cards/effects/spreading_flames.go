package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spreading Flames — Instant {6}{R} (Eldritch Moon, #5541):
//
//	"Spreading Flames deals 6 damage divided as you choose among any number
//	 of target creatures."
//
// Pyrokinesis's division at six (CR 601.2d: the split is announced with
// the targets, at least 1 to each; DealDividedDamage reads it back and
// skips a target that left).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95bedd11-ed01-4cb1-bcdb-c5f9fdba8c79",
		Name:         "Spreading Flames",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(Divide(6)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
