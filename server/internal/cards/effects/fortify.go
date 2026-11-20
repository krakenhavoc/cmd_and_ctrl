package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fortify — Instant {1}{W} (Time Spiral, #5596):
//
//	"Choose one —
//	 • Creatures you control get +2/+0 until end of turn.
//	 • Creatures you control get +0/+2 until end of turn."
//
// Borrowed Grace's two bullets without escalate: one mode, chosen as it
// is cast. The creatures are the ones you control as it resolves (CR
// 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52ea24d4-d33d-40fc-8ffc-09ae625b908f",
		Name:         "Fortify",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Creatures you control get +2/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Label: "Fortify — +2/+0"}.Apply(ctx)
				}),
			ModeDoing("Creatures you control get +0/+2 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl()), Toughness: 2, Label: "Fortify — +0/+2"}.Apply(ctx)
				}),
		),
	})
}
