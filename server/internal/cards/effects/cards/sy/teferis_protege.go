package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Protege — Creature — Human Wizard {2}{U}, 2/3 (Core Set 2021, #5511):
//
//	"{1}{U}, {T}: Draw a card, then discard a card."
//
// Erratic Visionary's loot: the draw first, then the discard prompt is
// built from the post-draw hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e84f1727-ce01-44be-aebf-8a0c4e032c9a",
		Name:         "Teferi's Protege",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
