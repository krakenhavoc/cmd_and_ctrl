package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx's Insight — Instant {2}{W}{U} (Ravnica Allegiance):
//
//	"Draw two cards.
//	 Addendum — If you cast this spell during your main phase, you gain 2
//	 life."
//
// Addendum is read as the spell resolves: it is the caster's own main phase,
// which is also when it was cast (Arrester's Zeal).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01f675c3-d1c4-45c8-8a2e-75f535db3025",
		Name:         "Sphinx's Insight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 2}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				if !effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
					return nil
				}
				return effects.GainLife{Amount: 2}.Apply(ctx)
			}),
		),
	})
}
