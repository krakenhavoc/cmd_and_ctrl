package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Romantic Rendezvous — Sorcery {1}{R} (Marvel Super Heroes Commander,
// #4455): "Discard a card, then draw two cards." The discard is the
// controller's choice and the draw waits for it (a registered
// continuation, as Charming Scoundrel's rummage); with an empty hand
// nothing is discarded and the two cards are still drawn (CR 701.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ca585b5-25c4-41cc-a0b9-dc66ae803d04",
		Name:         "Romantic Rendezvous",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: item.Controller,
				Source: item.SourceCardID,
				N:      1,
			}, mscM1RendezvousDraw, nil)
		},
	})
}

// mscM1RendezvousDraw is "then draw two cards", keyed: a table waiting
// on the discard is a restore point.
var mscM1RendezvousDraw = game.RegisterDiscardThen("choose-cards/msc-m1-romantic-rendezvous-draw", func(g *game.Game, r game.DiscardsDone) error {
	return g.DrawNForEffect(r.Player, 2)
})
