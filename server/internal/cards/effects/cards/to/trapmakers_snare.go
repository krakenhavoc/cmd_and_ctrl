package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trapmaker's Snare — Instant {1}{U} (Zendikar, #5583):
//
//	"Search your library for a Trap card, reveal it, put it into your hand,
//	 then shuffle."
//
// A plain tutor with a subtype test (Trap is a subtype of instants).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8aa0125c-29fe-4f31-985c-de2da3dbc602",
		Name:         "Trapmaker's Snare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.HasSubtype("Trap") },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Source:    ctx.Source(),
				Reason:    "Trapmaker's Snare — a Trap card",
			}.Apply(ctx)
		},
	})
}
