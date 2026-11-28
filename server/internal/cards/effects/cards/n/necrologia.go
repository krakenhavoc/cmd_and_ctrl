package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrologia — Instant {3}{B}{B} (Exodus, S82 OLD batch 742, #5626):
//
//	"Cast this spell only during your end step.
//	 As an additional cost to cast this spell, pay X life.
//	 Draw X cards."
//
// Toxic Deluge's PayXLifeCost: X is announced with the cast, the life is paid
// as a cost with the spell on the stack, and the draw reads the same X. The
// timing is a cast condition, true only in the end step of the caster's own
// turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "925f11c3-fd32-4aa7-a130-6554b81ad3eb",
		Name:           "Necrologia",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.PayXLifeCost(),
		CastCondition: func(g *game.Game, caster uuid.UUID, _ game.Card) bool {
			return g.Turn.Step == game.StepEnd && effects.IsYourTurn(g, caster)
		},
		CastConditionLabel: "Cast this spell only during your end step.",
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return effects.DrawCards{N: x}.Apply(ctx)
		},
	})
}
