package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twist Reality — Instant {1}{U}{U}:
//
//	"Choose one —
//	 • Counter target spell.
//	 • Manifest dread."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55b0518e-9c87-4f0e-81c2-9c984f760be1",
		Name:         "Twist Reality",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Counter target spell.", effects.TargetSpell("target spell")),
			effects.Mode("Manifest dread."),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(1) {
				return effects.ManifestDread{}.Apply(ctx)
			}
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
