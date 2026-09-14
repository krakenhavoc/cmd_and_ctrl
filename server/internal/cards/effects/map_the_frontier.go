package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Map the Frontier — Sorcery {3}{G} (EDHREC rank 3629):
//
//	"Search your library for up to two basic land cards and/or Desert
//	 cards, put them onto the battlefield tapped, then shuffle."
//
// The Desert deck's Explosive Vegetation. One search, one prompt:
// every basic land card and every card with the Desert land type is
// offered — a nonbasic Desert included, a Snow-Covered basic
// included — and up to two are taken, tapped, then the library is
// shuffled. The "tapped" is the search's own flag, so a Desert with
// an enters-tapped clause of its own enters tapped once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11a46cb6-ab01-4630-9541-782db2ef3b91",
		Name:         "Map the Frontier",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b34SearchUpToTwoBasicsOrDesertsTapped(ctx)
		},
	})
}
