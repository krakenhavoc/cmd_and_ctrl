package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Andromedan Automatons — Artifact Creature — Android {2}{W}, 3/1
// (Star Trek, #4973):
//
//	"When this creature enters, you may search your library for any
//	 number of cards named Andromedan Automatons, reveal them, put them
//	 into your hand, then shuffle."
//
// "Any number" is SearchLibrary's Unbounded pick (zero up to every
// copy), and "you may" is Optional, so the controller can decline the
// search and the shuffle. The name is read through the card's printed
// name, the way "named Shadowborn Apostle" is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2da322e6-a8e2-4d7a-bd34-54a3dfa85940",
		Name:         "Andromedan Automatons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Andromedan Automatons — search for cards named Andromedan Automatons",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.Name == "Andromedan Automatons" },
						Dest:      game.ZoneHand,
						Unbounded: true,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Andromedan Automatons — you may search for any number of cards named Andromedan Automatons",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
