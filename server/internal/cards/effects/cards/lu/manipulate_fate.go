package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manipulate Fate — Sorcery {1}{U} (Invasion, S82 #5616):
//
//	"Search your library for three cards, exile them, then shuffle.
//	 Draw a card."
//
// An unconditional search, so it must find three (CR 701.23d), or every card
// if the library holds fewer. The search can pause on a prompt, so the draw is
// the search's tail and runs once the cards are exiled and the library
// shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "541f13aa-a1ed-453a-b719-82daef1be607",
		Name:         "Manipulate Fate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:   item.Controller,
				Dest:     game.ZoneExile,
				Limit:    3,
				Shuffle:  true,
				Reason:   "Manipulate Fate — search your library for three cards to exile",
				Source:   item.SourceCardID,
				ThenTail: ctx.ItemTail(manipulateFateSearched),
			}.Apply(ctx)
		},
	})
}

// manipulateFateSearched is "Draw a card", after the search. An on-disk key.
var manipulateFateSearched = effects.CardTail("then/manipulate-fate-searched", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DrawCards{N: 1}.Apply(ctx)
})
