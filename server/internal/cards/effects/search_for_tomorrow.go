package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Search for Tomorrow — Sorcery {2}{G}:
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield, then shuffle.
//	 Suspend 2—{G}"
//
// Rampant Growth's search with the land untapped, and the suspend
// special action (CR 702.62) of Ancestral Vision.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cdc9f99-c6fa-40cd-90b0-e47d43a8cc3c",
		Name:         "Search for Tomorrow",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		SpecialActions: []game.SpecialAction{
			Suspend(2, "{G}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Search for Tomorrow — a basic land",
			}.Apply(ctx)
		},
	})
}
