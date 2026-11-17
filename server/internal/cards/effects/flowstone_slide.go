package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flowstone Slide — Sorcery {X}{2}{R}{R}
// (Nemesis, #5618):
//
//	"All creatures get +X/-X until end of turn."
//
// Toxic Deluge's body at +X/-X: every creature on the battlefield as
// the spell resolves (CR 611.2c), until the cleanup step (CR 514.2). A
// creature whose toughness falls to 0 or less dies to the state-based
// check (CR 704.5f). X is 0 when the spell is not cast for one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad779760-1197-4248-8e21-7a048122640f",
		Name:         "Flowstone Slide",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return BoostUntilEOT{Match: Creature(), Power: x, Toughness: -x, Label: "Flowstone Slide — +X/-X"}.Apply(ctx)
		},
	})
}
