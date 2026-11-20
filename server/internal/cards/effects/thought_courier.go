package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Courier — Creature — Human Wizard {1}{U}, 1/1 (Fifth Dawn, #5604):
//
//	"{T}: Draw a card, then discard a card."
//
// Merfolk Looter's loot: draw first, then the discard prompt (CR 701.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06ea02f5-f296-4334-bb6f-9cb018e1e150",
		Name:         "Thought Courier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
		}},
	})
}
