package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pondering Mage — Creature — Human Wizard {3}{U}{U}, 3/4 (Modern Horizons):
//
//	"When this creature enters, look at the top three cards of your library,
//	 then put them back in any order. You may shuffle. Draw a card."
//
// Ponder's chain as an enters trigger: the order is asked first, the shuffle
// second and the draw follows on both answers, so a shuffle throws the order
// away and the card drawn is the one on top after it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f278970-201d-4a93-b038-cc7218f2ac57",
		Name:         "Pondering Mage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pondering Mage — look at the top three cards, put them back in any order, you may shuffle, then draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.OldB159LookThreeMayShuffleThenDraw(effects.NewContext(g, item))
				}),
		},
	})
}
