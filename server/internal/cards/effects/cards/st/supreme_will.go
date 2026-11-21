package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Supreme Will — Instant {2}{U} (Foundations Jumpstart, #5929):
//
//	"Choose one —
//	 • Counter target spell unless its controller pays {3}.
//	 • Look at the top four cards of your library. Put one of them into
//	   your hand and the rest on the bottom of your library in any order."
//
// The first bullet is CounterUnlessPaid (CR 118.12) and the second is
// Impulse's sentence, LookAtTopThenTakeOneRestOnBottom.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75325029-e6f1-44af-88b2-53cc2b7233c0",
		Name:         "Supreme Will",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Counter target spell unless its controller pays {3}.", effects.TargetSpell("target spell")),
			effects.ModeWithPurpose(effects.ModeDoing("Look at the top four cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.LookAtTopThenTakeOneRestOnBottom(4, "Supreme Will — put one of those cards into your hand")(ctx.Game, item)
				}), game.Purpose{Draws: 1}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if !ctx.HasMode(0) {
				return nil
			}
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     "{3}",
				Question: "Supreme Will — pay {3} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
