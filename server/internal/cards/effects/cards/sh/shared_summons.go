package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shared Summons — Instant {3}{G}{G} (EDHREC rank 3986):
//
//	"Search your library for up to two creature cards with different
//	 names, reveal them, put them into your hand, then shuffle."
//
// Five mana at instant speed for two creatures, which in a
// creature-combo deck is both halves of the combo in one card. It is
// in the batch as the "different names" search constraint — a rule
// about the picked SET that no per-card predicate can express — and
// Tiamat is the only other card in the catalog that carries one.
//
// The three clauses each do work:
//
//   - "UP TO two" makes it legal to find one, or none. A deck with
//     one relevant creature left still casts it.
//   - "with different names" is checked over the whole pick, not per
//     card, so two copies of the same creature are not a legal
//     answer — the reason the constraint rides Validate rather than
//     the predicate.
//   - "reveal them" makes the picks public, so the table sees the
//     combo coming. Search hidden-information discipline (only the
//     searcher sees the library) is the engine's, not this file's.
//
// "Creature card" is read off the PRINTED type line, as every library
// search is: a card in a library has no layered characteristics to
// read, so a changeling-granting effect on the battlefield does not
// widen what this can find.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2c9dbd0-2062-4ee1-b32e-8eeacd95589c",
		Name:         "Shared Summons",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: game.Card.IsCreature,
				Dest:      game.ZoneHand,
				Limit:     2,
				Reveal:    true,
				Shuffle:   true,
				Rule:      effects.SearchRuleDifferentNames,
				Reason:    "Shared Summons — up to two creature cards with different names, revealed, to hand",
			}.Apply(ctx)
		},
	})
}
