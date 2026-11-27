package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gaea's Bounty — Sorcery {2}{G} (Urza's Saga, S82 OLD b722):
//
//	Search your library for up to two Forest cards, reveal those cards, put them into your hand, then shuffle.
//
// A search with a stated quality may find fewer than two (CR 701.23b); the cards are revealed and shuffled after. "Forest card" is a land with the subtype Forest, basic or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5a61e61-d5fe-41ef-afd6-ba5e0399b986",
		Name:         "Gaea's Bounty",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsLandWithSubtype("forest"),
				Dest:      game.ZoneHand,
				Limit:     2,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Gaea's Bounty — up to two Forest cards",
			}.Apply(ctx)
		},
	})
}
