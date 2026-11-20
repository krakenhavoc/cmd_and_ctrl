package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seek the Horizon — Sorcery {3}{G} (Saviors of Kamigawa, #5601):
//
//	"Search your library for up to three basic land cards, reveal them,
//	 put them into your hand, then shuffle."
//
// Land Tax's search, cast as a sorcery: the search has a stated quality,
// so it may find fewer than three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7537b482-7626-4b65-8b67-53b99baf1872",
		Name:         "Seek the Horizon",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 3},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     3,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Seek the Horizon — up to three basic land cards, to your hand",
			}.Apply(ctx)
		},
	})
}
