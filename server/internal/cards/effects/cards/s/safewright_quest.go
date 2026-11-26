package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Safewright Quest — Sorcery {G/W} (Shadowmoor, #5591):
//
//	"Search your library for a Forest or Plains card, reveal it, put it
//	 into your hand, then shuffle."
//
// A Forest or Plains card is any card with that land type, basic or not
// (Gift of Estates' reading).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "003c48d4-0303-4970-8637-47ae060385aa",
		Name:         "Safewright Quest",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					return c.HasSubtypeWord("Forest") || c.HasSubtypeWord("Plains")
				},
				Dest:    game.ZoneHand,
				Limit:   1,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Safewright Quest — a Forest or Plains card",
			}.Apply(effects.NewContext(ctx.Game, item))
		},
	})
}
