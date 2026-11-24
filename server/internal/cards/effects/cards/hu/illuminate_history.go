package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illuminate History — Sorcery — Lesson {2}{R}{R} (Strixhaven: School of Mages,
// #5502):
//
//	"Discard any number of cards, then draw that many cards. Then if there are
//	 seven or more cards in your graveyard, create a 3/2 red and white Spirit
//	 creature token."
//
// "Any number" is a real choice, from nothing to the whole hand, and the rest is
// the discard's keyed continuation (Steal the Show's shape): draw as many cards
// as were really discarded (CR 701.9a), then count the graveyard. The spell is
// still on the stack, so it is not among the seven. Being a Lesson is its type
// line, which learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7c1e7cf-0be2-4129-9bdf-0862be2fba7f",
		Name:         "Illuminate History",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.B14HandSize(ctx.Game, ctx.Controller())
			if n == 0 {
				return illuminateHistorySpirit(ctx)
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   ctx.Controller(),
				Source:   item.SourceCardID,
				N:        n,
				UpTo:     true,
				Question: "Illuminate History — discard any number of cards, then draw that many",
			}, illuminateHistoryDraw, nil)
		},
	})
}

// illuminateHistoryDraw is "then draw that many cards. Then if there are seven or
// more cards in your graveyard, create a 3/2 Spirit", keyed.
var illuminateHistoryDraw = game.RegisterDiscardThen("choose-cards/s82-old-b046-illuminate-history", func(g *game.Game, r game.DiscardsDone) error {
	if err := g.DrawNForEffect(r.Player, r.Discarded.Count()); err != nil {
		return err
	}
	return illuminateHistorySpirit(effects.KeyedContext(g, r.Player, r.Source))
})

// illuminateHistorySpirit makes the Spirit when the player's graveyard holds
// seven or more cards.
func illuminateHistorySpirit(ctx *effects.Context) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Graveyard == nil || p.Graveyard.Size() < 7 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("3/2 red and white Spirit"), N: 1}.Apply(ctx)
}
