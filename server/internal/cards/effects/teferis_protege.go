package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi's Protege — Creature — Human Wizard {2}{U}, 2/3 (Core Set 2021, #5511):
//
//	"{1}{U}, {T}: Draw a card, then discard a card."
//
// Erratic Visionary's loot: the draw first, then the discard prompt is
// built from the post-draw hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e84f1727-ce01-44be-aebf-8a0c4e032c9a",
		Name:         "Teferi's Protege",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
	})
}
