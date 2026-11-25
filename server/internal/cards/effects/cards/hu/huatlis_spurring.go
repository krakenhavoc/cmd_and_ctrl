package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huatli's Spurring — Instant {R} (Ixalan):
//
//	"Target creature gets +2/+0 until end of turn. If you control a Huatli
//	 planeswalker, that creature gets +4/+0 until end of turn instead."
//
// One target and one pump, either +2/+0 or +4/+0: whether its controller
// controls a Huatli planeswalker is read as the spell resolves, and the
// target is re-checked then (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bba7c7d-6110-4d7f-8b28-3c5b9d3448df",
		Name:         "Huatli's Spurring",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			power := 2
			if effects.OldB222YouControlAPlaneswalker(ctx.Game, ctx.Controller(), "Huatli") {
				power = 4
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: power, Label: "Huatli's Spurring — +X/+0"}.Apply(ctx)
		},
	})
}
