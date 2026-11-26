package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selective Memory — Sorcery {3}{U} (Worldwake, S82 OLD b488):
//
//	"Search your library for any number of nonland cards, exile them, then
//	 shuffle."
//
// A search with no limit (Sarkhan, Dragonsoul's "any number"): you choose which of the
// nonland cards to take, any number of them from none to all (a search that names a quality
// may find nothing, CR 701.23b), they are exiled face up, and the library is shuffled
// whether or not anything was taken. Nothing may be cast from exile afterwards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cfb4977-5b6d-4220-9c54-70920b3087ed",
		Name:         "Selective Memory",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return !c.IsLand() },
				Dest:      game.ZoneExile,
				Unbounded: true,
				Shuffle:   true,
				Reason:    "Selective Memory — any number of nonland cards, exile them",
			}.Apply(ctx)
		},
	})
}
