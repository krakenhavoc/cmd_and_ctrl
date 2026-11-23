package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "7e484372-a052-4df6-a6b2-2c6accccabef",
		Name:         "Wretched Throng",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Wretched Throng — search your library for a card named Wretched Throng",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.Name == "Wretched Throng" },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Wretched Throng — you may search for a card named Wretched Throng",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
