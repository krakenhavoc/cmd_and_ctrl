package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fires of Victory — Instant {1}{R} (Dominaria United):
//
//	"Kicker {2}{U} (You may pay an additional {2}{U} as you cast this
//	 spell.)
//	 If this spell was kicked, draw a card.
//	 Fires of Victory deals damage to target creature or planeswalker
//	 equal to the number of cards in your hand."
//
// Printed order is resolution order (CR 608.2c): the kicked draw comes
// first, so the card it draws counts. The hand is counted once as the
// damage is dealt; the spell is on the stack then, not in hand. If the
// target is gone the spell does nothing, drawing included (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d6afb1a9-c9ed-4a56-ba7d-700ab71f2e0d",
		Name:          "Fires of Victory",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{U}")},
		Targets:       effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.DrawCards{N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.DealDamage{
					Target: t.ID,
					Amount: effects.B14HandSize(ctx.Game, ctx.Controller()),
					Source: ctx.Source(),
				}.Apply(ctx)
			}),
		),
	})
}
