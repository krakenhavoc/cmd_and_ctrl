package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mysteries of the Deep — Instant {4}{U} (Worldwake, #5582):
//
//	"Draw two cards.
//	 Landfall — If you had a land enter the battlefield under your control
//	 this turn, draw three cards instead."
//
// "If you had a land enter the battlefield under your control this
// turn" reads the controller's turn tally (wwkHadLandfallThisTurn),
// which counts every land that entered under them this turn, one that
// has since left included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3708eaae-9306-46b2-ae41-7ce6c50be3c7",
		Name:         "Mysteries of the Deep",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 2
			if effects.WwkHadLandfallThisTurn(ctx) {
				n = 3
			}
			return effects.DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
		},
	})
}
