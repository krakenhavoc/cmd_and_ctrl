package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Erratic Visionary — Creature — Human Wizard {1}{U}, 1/3 (Ravnica: Clue
// Edition):
//
//	"{1}{U}, {T}: Draw a card, then discard a card."
//
// Agna Qel'a's loot: the draw first, then the discard prompt is built from
// the post-draw hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "334c0913-0ce0-4934-979d-64dbc1ec4aa2",
		Name:         "Erratic Visionary",
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
