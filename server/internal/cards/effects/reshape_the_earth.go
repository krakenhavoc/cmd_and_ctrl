package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reshape the Earth — Sorcery {6}{G}{G}{G} (EDHREC rank 3358):
//
//	"Search your library for up to ten land cards, put them onto the
//	 battlefield tapped, then shuffle."
//
// Nine mana for the whole Field of the Dead package. One search with
// a limit of ten and the "up to" as a real prompt the searcher may
// take fewer from; every land arrives tapped through the fetching
// effect's own flag, each running its own entry pipeline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edc15ea8-d321-4884-bdcf-ae6198aab78b",
		Name:         "Reshape the Earth",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     func(c game.Card) bool { return c.IsLand() },
				Dest:          game.ZoneBattlefield,
				Limit:         10,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Optional:      true,
				Reason:        "Reshape the Earth — choose up to ten land cards to put onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
