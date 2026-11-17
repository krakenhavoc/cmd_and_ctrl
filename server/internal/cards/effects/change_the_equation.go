package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Change the Equation — Instant {1}{U} (March of the Machine, S82 #5477):
//
//	"Choose one —
//	 • Counter target spell with mana value 2 or less.
//	 • Counter target red or green spell with mana value 6 or less."
//
// Each mode names its own target statement, so the chosen mode alone
// constrains the target (CR 700.2); the mana value of a spell on the
// stack is read as it is now (CR 202.3). Colour is read through the
// spell's effective colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aba7e9e2-8df8-454e-a332-9abae1d95b6f",
		Name:         "Change the Equation",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Counter target spell with mana value 2 or less.", TargetSpell("target spell with mana value 2 or less", ManaValueLE(2))),
			Mode("Counter target red or green spell with mana value 6 or less.",
				TargetSpell("target red or green spell with mana value 6 or less", And(Or(OfColor("R"), OfColor("G")), ManaValueLE(6)))),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				return CounterTarget{StackID: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
