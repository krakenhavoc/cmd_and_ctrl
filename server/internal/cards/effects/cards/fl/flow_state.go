package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flow State — Sorcery {1}{U} (Secrets of Strixhaven, #4456):
//
//	"Look at the top three cards of your library. Put one of them into your
//	 hand and the rest on the bottom of your library in any order. If there
//	 is an instant card and a sorcery card in your graveyard, instead put two
//	 of them into your hand and the rest on the bottom of your library in any
//	 order."
//
// Impulse's dig with a count that is read as the spell resolves (CR 608.2c):
// two when the graveyard holds an instant card and a sorcery card, one
// otherwise. "Put one of them" is mandatory, so the take has no "may"; a
// library of fewer cards is taken with no prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7194bca-0cf4-4a86-8b84-d2f9219a08b3",
		Name:         "Flow State",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			take, label := 1, "Flow State — put one of them into your hand"
			if effects.SocB25GraveyardHasInstantAndSorcery(ctx.Game, player) {
				take, label = 2, "Flow State — put two of them into your hand"
			}
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    take,
				Label:  label,
				Key:    effects.TakeRestBottomAny,
			}.Apply(ctx)
		},
	})
}
