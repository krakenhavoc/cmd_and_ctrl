package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beneath the Sands — Sorcery {2}{G} (Hour of Devastation, #5534):
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Shared Roots' search, and the shared Cycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27547519-9474-49fa-80e7-e604d61cb3ca",
		Name:         "Beneath the Sands",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Beneath the Sands — a basic land onto the battlefield tapped",
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
