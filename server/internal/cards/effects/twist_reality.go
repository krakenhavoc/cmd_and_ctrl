package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twist Reality — Instant {1}{U}{U}:
//
//	"Choose one —
//	 • Counter target spell.
//	 • Manifest dread."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55b0518e-9c87-4f0e-81c2-9c984f760be1",
		Name:         "Twist Reality",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Counter target spell.", TargetSpell("target spell")),
			Mode("Manifest dread."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(1) {
				return ManifestDread{}.Apply(ctx)
			}
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
