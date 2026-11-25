package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Call the Gatewatch — Sorcery {2}{W} (Oath of the Gatewatch, #5545):
//
//	"Search your library for a planeswalker card, reveal it, put it into your
//	 hand, then shuffle."
//
// The search is not optional, but the searcher may find nothing, and the
// library is shuffled either way (Arena Rector's planeswalker search, to the
// hand and revealed).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "652bfa31-b2eb-4cba-850f-1db02a83bb35",
		Name:         "Call the Gatewatch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsPlaneswalker() },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Call the Gatewatch — a planeswalker card, revealed, to hand",
			}.Apply(ctx)
		},
	})
}
