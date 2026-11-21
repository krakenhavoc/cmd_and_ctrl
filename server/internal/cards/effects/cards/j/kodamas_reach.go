package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kodama's Reach — Sorcery — Arcane {2}{G}:
//
//	"Search your library for up to two basic land cards, reveal those
//	cards, put one onto the battlefield tapped and the other into
//	your hand, then shuffle."
//
// Functionally identical to Cultivate (the Arcane subtype matters
// only to splice, which reads the type line, ADR 0202), so it shares
// cultivateSearch: one optional search per destination, the first
// deferring its shuffle, the second shuffling once. "Up to two" is
// honoured (#2942) — see cultivate.go.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90",
		Name:         "Kodama's Reach",
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CultivateSearch(ctx, "Kodama's Reach")
		},
	})
}
