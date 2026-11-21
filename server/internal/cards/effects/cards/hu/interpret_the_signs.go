package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Interpret the Signs — Sorcery {5}{U} (Jumpstart 2022, #5948):
//
//	"Scry 3, then reveal the top card of your library. Draw cards equal to
//	 that card's mana value."
//
// Everything after "then" is the scry's continuation (Scry.ThenTail), so the
// card revealed is the one the player left on top. The mana value is read off
// the revealed card as it is revealed; an empty library reveals nothing and
// draws nothing. The draw is as many cards as the mana value, one at a time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "834554f6-3b65-44fd-b1d1-8495b740dbfc",
		Name:         "Interpret the Signs",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{Player: item.Controller, N: 3, ThenTail: ctx.ItemTail(interpretTheSignsRevealed)}.Apply(ctx)
		},
	})
}

// interpretTheSignsRevealed is "reveal the top card of your library. Draw
// cards equal to that card's mana value". An on-disk key: never renamed,
// never reused.
var interpretTheSignsRevealed = effects.CardTail("interpret-the-signs/reveal-and-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	var ids []uuid.UUID
	if err := (effects.RevealTopOfLibrary{N: 1, Reason: "Interpret the Signs", Revealed: &ids}).Apply(ctx); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(ids[0])
	if !ok {
		return nil
	}
	return effects.DrawCards{N: c.ManaValue()}.Apply(ctx)
})
