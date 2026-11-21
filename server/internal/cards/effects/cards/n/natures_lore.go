package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nature's Lore — Sorcery {1}{G}:
//
//	"Search your library for a Forest card, put it onto the
//	battlefield, then shuffle."
//
// UNTAPPED is the entire point — this and Three Visits are premium
// over Rampant Growth because the land is usable the turn it lands,
// so TappedOnEntry is deliberately false here.
//
// "A Forest card" means any land with the Forest subtype, so
// IsLandWithSubtype rather than IsBasicLand: a Snow-Covered Forest
// or a Bayou both qualify in paper and both qualify here.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78826359-fe63-44ad-adc4-a17ffcd710e4",
		Name:         "Nature's Lore",
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsLandWithSubtype("forest"),
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Nature's Lore — a Forest card",
			}.Apply(ctx)
		},
	})
}
