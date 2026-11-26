package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fathom Trawl — Sorcery {3}{U}{U} (Lorwyn, S82 OLD b546, #5593):
//
//	"Reveal cards from the top of your library until you reveal three
//	 nonland cards. Put the nonland cards revealed this way into your hand,
//	 then put the rest of the revealed cards on the bottom of your library
//	 in any order."
//
// The run is revealed as one reveal and stops at the third nonland card (or
// the end of the library). Every nonland card revealed goes to hand and the
// rest goes to the bottom in the order the player chooses (TakeRestBottomAny).
//
// Purpose: it is a draw engine of three cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b002cae-ee7d-400d-af9c-445eb10f1f0c",
		Name:         "Fathom Trawl",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			run, _ := effects.RevealUntilCount(ctx, player, func(c game.Card) bool { return !c.IsLand() }, 3, "Fathom Trawl — reveal")
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  run,
				Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !c.IsLand() },
				All:    true,
				Label:  "Fathom Trawl — put the nonland cards revealed this way into your hand",
				Key:    effects.TakeRestBottomAny,
			}.Apply(ctx)
		},
	})
}
