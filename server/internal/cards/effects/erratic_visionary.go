package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "334c0913-0ce0-4934-979d-64dbc1ec4aa2",
		Name:         "Erratic Visionary",
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
