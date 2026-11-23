package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Price — Sorcery {3}{B} (Jumpstart 2022, #5948):
//
//	"Look at the top four cards of your library. Put two of them into
//	 your hand and the rest on the bottom of your library in any order.
//	 You lose 2 life."
//
// The two-card take is mandatory (no "may"): the prompt's floor is the
// ceiling. The life is lost after the take is set up, as the text orders it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93c6554f-096d-4556-81ca-e87e29265887",
		Name:         "Blood Price",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			if err := (effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Max:    2,
				Label:  "Blood Price — put two of those cards into your hand and the rest on the bottom of your library",
				Key:    effects.TakeRestBottomAny,
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: player, Amount: -2}.Apply(ctx)
		},
	})
}
