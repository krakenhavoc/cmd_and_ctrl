package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Three Visits — Sorcery {1}{G}:
//
//	"Search your library for a Forest card, put it onto the
//	battlefield, then shuffle."
//
// Functionally identical to Nature's Lore, and played alongside it
// for exactly that reason. Untapped, so TappedOnEntry stays false.
func init() {
	Register(Spec{
		OracleID:     "1b882a0e-0ede-4d1a-bd1a-9b7cffbcde8e",
		Name:         "Three Visits",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: IsLandWithSubtype("forest"),
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Three Visits — a Forest card",
			}.Apply(ctx)
		},
	})
}
