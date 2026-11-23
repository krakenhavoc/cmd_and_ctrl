package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stress Dream — Instant {3}{U}{R} (Secrets of Strixhaven, #4456):
//
//	"Stress Dream deals 5 damage to up to one target creature. Look at the
//	 top two cards of your library. Put one of those cards into your hand and
//	 the other on the bottom of your library."
//
// "Up to one" lets the spell be cast with no target; the look-and-take does
// not depend on the damage, so it happens either way. It is its own step
// behind the damage (CR 608.2c), which can pause on a replacement order.
// Impulse's look: private, a mandatory pick of exactly one, the other on the
// bottom (a pile of one has no order to ask for).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52124954-b0b6-492c-8d03-fe54c220facc",
		Name:         "Stress Dream",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(0); ok {
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.LookAtTopThenTakeOneRestOnBottom(2,
					"Stress Dream — put one of those cards into your hand and the other on the bottom of your library")(ctx.Game, ctx.Item)
			}),
		),
	})
}
