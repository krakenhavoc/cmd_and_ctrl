package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abuna's Chant — Instant {3}{W}:
//
//	"Choose one —
//	 • You gain 5 life.
//	 • Prevent the next 5 damage that would be dealt to target creature
//	   this turn.
//	 Entwine {2} (Choose both if you pay the entwine cost.)"
//
// Entwine (CR 702.42, #3866): {2} more takes both bullets. The shield is
// Mending Hands' charged PreventNextDamage, five points, until end of
// turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "2d4721f1-1a49-45aa-99d5-ab8c6f3f3ab6",
		Name:          "Abuna's Chant",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("You gain 5 life.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GainLife{Amount: 5}.Apply(ctx)
				}),
			effects.ModeDoing("Prevent the next 5 damage that would be dealt to target creature this turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.PreventNextDamage{
						Target: t.ID,
						Amount: 5,
						Label:  "Abuna's Chant: prevent the next 5 damage",
					}.Apply(ctx)
				}),
		)),
	})
}
