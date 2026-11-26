package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sands of Delirium — Artifact {3} (Magic 2013, S82 OLD, #5568):
//
//	"{X}, {T}: Target player mills X cards."
//
// X is the ability's own, paid in its cost (CR 107.3k) and read as it resolves.
// A target that is gone does nothing (CR 608.2b); X of zero mills nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2cc67bf-59c5-4e06-93a3-dcaa1bf276ba",
		Name:         "Sands of Delirium",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Target player mills X cards.",
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if p := effects.TargetedPlayerOf(ctx); p != [16]byte{} && ctx.X() > 0 {
					return effects.MillCards{Player: p, N: ctx.X()}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
