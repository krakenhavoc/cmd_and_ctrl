package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unmarked Grave — Sorcery {1}{B} (EDHREC rank 1610):
//
//	"Search your library for a nonlegendary card, put that card into
//	 your graveyard, then shuffle."
//
// Entomb with a legendary clause. A library search to the graveyard
// — the searcher picks from every nonlegendary card in the library,
// so the prompt always opens, and "fail to find" is a legal answer
// (CR 701.19c). Legendary is read off the effective characteristic,
// which for a card in a library is the printed supertype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "134dfb35-0c32-4143-aa63-7701f406b59e",
		Name:         "Unmarked Grave",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return !c.IsLegendary() },
				Dest:      game.ZoneGraveyard,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Unmarked Grave — a nonlegendary card, to your graveyard",
			}.Apply(ctx)
		},
	})
}
