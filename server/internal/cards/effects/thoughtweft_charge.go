package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thoughtweft Charge — Instant {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"Target creature gets +3/+3 until end of turn. If a creature entered
//	 the battlefield under your control this turn, draw a card."
//
// If the target is illegal on resolution nothing happens and nothing is
// drawn (the ruling): the engine removes a spell whose only target is gone.
func init() {
	Register(Spec{
		OracleID:     "a8cdcaa0-8e25-4c17-b5ce-16f72b535123",
		Name:         "Thoughtweft Charge",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump, Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 3, 3))},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := vsPumpTheTarget(ctx, 3, 3, "Thoughtweft Charge — +3/+3 until end of turn"); err != nil {
				return err
			}
			if ctx.Game.EnteredWithCardTypeThisTurn(item.Controller, "creature") == 0 {
				return nil
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
