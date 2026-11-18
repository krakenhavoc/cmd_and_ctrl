package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wretched Throng — Creature — Zombie Horror {1}{U}, 2/1 (Innistrad: Crimson
// Vow):
//
//	"When this creature dies, you may search your library for a card named
//	 Wretched Throng, reveal it, put it into your hand, then shuffle."
//
// A "you may search" is SearchLibrary's Optional, so the controller can
// decline the search and the shuffle. The name is the card's printed name, as
// Andromedan Automatons reads it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e484372-a052-4df6-a6b2-2c6accccabef",
		Name:         "Wretched Throng",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Wretched Throng — search your library for a card named Wretched Throng",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.Name == "Wretched Throng" },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Wretched Throng — you may search for a card named Wretched Throng",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
