package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spoils of Victory — Sorcery {2}{G} (Commander 2013, S82 OLD b380, #5560):
//
//	"Search your library for a Plains, Island, Swamp, Mountain, or Forest card
//	 and put that card onto the battlefield. Then shuffle."
//
// A search over lands with a basic land type (a nonbasic dual qualifies,
// CR 205.3i), put onto the battlefield untapped, nothing revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "852bd598-6e48-43c8-9211-740ae9e0c42e",
		Name:         "Spoils of Victory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Source:    item.SourceCardID,
				Predicate: effects.HasABasicLandType,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Spoils of Victory — a Plains, Island, Swamp, Mountain, or Forest card onto the battlefield",
			}.Apply(ctx)
		},
	})
}
