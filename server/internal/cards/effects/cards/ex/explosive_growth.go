package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explosive Growth — Instant {G} (Invasion, S82 OLD #5616):
//
//	"Kicker {5} (You may pay an additional {5} as you cast this spell.)
//	 Target creature gets +2/+2 until end of turn. If this spell was kicked,
//	 that creature gets +5/+5 until end of turn instead."
//
// Might of Murasa's shape: one pump, its size read off the cast. The
// "instead" replaces the +2/+2, so a kicked cast is +5/+5 and never +7/+7.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "96c65faf-859c-4436-a551-5538409a9792",
		Name:          "Explosive Growth",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{5}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 2
			if ctx.WasKicked() {
				n = 5
			}
			return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Explosive Growth — pump"}.Apply(ctx)
		},
	})
}
