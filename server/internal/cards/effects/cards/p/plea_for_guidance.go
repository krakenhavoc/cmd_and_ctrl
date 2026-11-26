package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plea for Guidance — Sorcery {5}{W} (Born of the Gods, S82 OLD b378, #5559):
//
//	"Search your library for up to two enchantment cards, reveal them, put
//	 them into your hand, then shuffle."
//
// Yasharn's two-card search to hand with the shuffle, over enchantment cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "358cfcf3-8ad0-4877-a59a-76ee269169fb",
		Name:         "Plea for Guidance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Source:    item.SourceCardID,
				Predicate: func(c game.Card) bool { return c.IsEnchantment() },
				Dest:      game.ZoneHand, Limit: 2, Reveal: true, Shuffle: true,
				Reason: "Plea for Guidance — up to two enchantment cards",
			}.Apply(ctx)
		},
	})
}
