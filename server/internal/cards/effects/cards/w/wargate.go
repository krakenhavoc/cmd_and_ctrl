package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wargate — Sorcery {X}{G}{W}{U} (Alara Reborn, #5586):
//
//	"Search your library for a permanent card with mana value X or less, put
//	 it onto the battlefield, then shuffle."
//
// The search is Mind into Matter's "permanent card with mana value X or less"
// over the library, with the X the spell was cast with (CR 107.3f; a land is
// mana value 0). The search may find nothing (CR 701.23b), and the library is
// shuffled either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26f39abe-ea0b-400d-b11c-03ee4cf50915",
		Name:         "Wargate",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player: item.Controller,
				Predicate: func(c game.Card) bool {
					return effects.And(effects.Permanent(), effects.ManaValueLE(ctx.X()))(ctx.Game, item.Controller, c)
				},
				Dest:     game.ZoneBattlefield,
				Limit:    1,
				Shuffle:  true,
				Optional: true,
				Reason:   "Wargate — a permanent card with mana value X or less",
			}.Apply(ctx)
		},
	})
}
