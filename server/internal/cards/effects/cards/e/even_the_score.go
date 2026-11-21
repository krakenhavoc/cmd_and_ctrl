package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Even the Score — Instant {X}{U}{U}{U} (Streets of New Capenna, #5491):
//
//	"This spell costs {U}{U}{U} less to cast if an opponent has drawn four
//	 or more cards this turn.
//	 Draw X cards."
//
// The reduction takes three blue mana off the cost (CR 118.7), so with the
// condition met the spell costs {X} alone. It reads ONE opponent's own
// tally of cards drawn this turn, not the table's total. X is the announced
// X (CR 601.2b) and the draw is the caster's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25f5cca6-a414-46d9-82e3-ce2fa21b14ca",
		Name:         "Even the Score",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		SelfCostModifiers: []game.CostModifier{
			effects.S82SncB14CostsLessMana("{U}{U}{U}",
				"This spell costs {U}{U}{U} less to cast if an opponent has drawn four or more cards this turn.",
				effects.S82SncB14AnOpponentDrewAtLeast(4)),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: ctx.X()}.Apply(ctx)
		},
	})
}
