package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lay of the Land — Sorcery {G} (Conspiracy: Take the Crown, S82 older
// sets batch 271, #5540):
//
//	"Search your library for a basic land card, reveal it, put it into your
//	 hand, then shuffle."
//
// Bushwhack's search as a spell of its own: one basic land card, revealed
// and put into the hand, then the shuffle; a search for a stated quality
// may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cedc52eb-66a6-4b43-87f1-9bb9f4d4871e",
		Name:         "Lay of the Land",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Lay of the Land — a basic land card",
			}.Apply(ctx)
		},
	})
}
