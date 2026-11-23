package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Might of Murasa — Instant {1}{G} (Zendikar Rising, #5508):
//
//	"Kicker {2}{G} (You may pay an additional {2}{G} as you cast this
//	 spell.)
//	 Target creature gets +3/+3 until end of turn. If this spell was
//	 kicked, that creature gets +5/+5 until end of turn instead."
//
// Colossal Growth's shape: one pump, its size read off the cast
// (ctx.WasKicked). The "instead" replaces the +3/+3, so a kicked cast is
// +5/+5 and never +8/+8.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "902eb71f-62b3-40cb-816b-f6b7e92e91e7",
		Name:          "Might of Murasa",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 3
			if ctx.WasKicked() {
				n = 5
			}
			return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Might of Murasa — pump"}.Apply(ctx)
		},
	})
}
