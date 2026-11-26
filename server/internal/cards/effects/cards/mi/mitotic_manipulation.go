package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mitotic Manipulation — Sorcery {1}{U}{U} (Mirrodin Besieged, S82 OLD b465,
// #5576):
//
//	"Look at the top seven cards of your library. You may put one of those
//	 cards onto the battlefield if it has the same name as a permanent. Put
//	 the rest on the bottom of your library in any order."
//
// A LOOK (CR 701.20, only the caster learns the cards), then a put of at most
// one qualifying card: its name must equal the name of a permanent on the
// battlefield (any player's, tokens included), read once as the spell
// resolves. Nonpermanent cards can't be put onto the battlefield, so they are
// never offered. The rest go on the bottom in an order the player chooses
// (BnsB06PutRestOnBottomInAnyOrder).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69cc006d-172f-4e4c-bb75-10d772b061b1",
		Name:         "Mitotic Manipulation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			cards := ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 7)
			if len(cards) == 0 {
				return nil
			}
			names := map[string]bool{}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				names[c.Name] = true
			}
			return effects.PutFromLibraryOntoBattlefield{
				Player: item.Controller,
				Cards:  cards,
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return names[c.Name]
				},
				Max:      1,
				Optional: true,
				Label:    "Mitotic Manipulation — you may put a card with the same name as a permanent onto the battlefield",
				ThenTail: effects.BnsB06PutRestOnBottomInAnyOrder,
			}.Apply(ctx)
		},
	})
}
