package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cateran Summons — Sorcery {B} (Mercadian Masques, S82 OLD, #5619):
//
//	"Search your library for a Mercenary card, reveal that card, put it
//	 into your hand, then shuffle."
//
// Priest of the Wakening Sun's search for a creature type: any card with the
// Mercenary subtype (read through the spell's own text), and it may find
// nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "910ff092-7c9d-49d3-a6df-497683e45bbd",
		Name:         "Cateran Summons",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			mercenary := ctx.TextSubtype("Mercenary")
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.HasSubtype(mercenary) },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Cateran Summons — a Mercenary card",
			}.Apply(ctx)
		},
	})
}
