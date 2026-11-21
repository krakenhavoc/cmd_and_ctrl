package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thirst for Meaning — Instant {2}{U}:
//
//	"Draw three cards. Then discard two cards unless you discard an
//	 enchantment card."
//
// Thirst for Discovery with enchantment cards where that card has basic
// lands: three cards now, and the discard is either one enchantment card
// or any two. The draws land first, then the discard prompt opens over the
// whole hand, so the drawn cards are among the choices
// (discardTwoUnlessOne).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70e58068-90d2-4720-a11c-5d243046b4a3",
		Name:         "Thirst for Meaning",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
				Ref: effects.DrawThenDiscardTwoUnlessEnchantment, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}
