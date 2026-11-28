package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unfulfilled Desires — Enchantment {1}{U}{B} (Mirage):
//
//	"{1}, Pay 1 life: Draw a card, then discard a card."
//
// A loot with a mana-and-life cost (CR 119.4, CR 602.2): the life is paid as
// the ability is activated, and the discard follows the draw as one effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5827d57-ae8b-44b0-a81b-3d191a0ba2fa",
		Name:         "Unfulfilled Desires",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Pay 1 life: Draw a card, then discard a card.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.PayLife(1)),
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
