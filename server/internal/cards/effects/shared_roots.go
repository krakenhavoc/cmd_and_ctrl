package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shared Roots — Sorcery — Lesson {1}{G} (EDHREC rank 1094):
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle."
//
// Rampant Growth with a Lesson subtype. The subtype matters only to
// "learn", which no card in the catalog has; the spell itself is the
// two-mana basic fetch, and the searcher picks the basic.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e0fd3bf-f47a-4f06-8ff1-73f6bf5d1e03",
		Name:         "Shared Roots",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Shared Roots — a basic land onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
