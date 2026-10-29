package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain of Umbar — Creature — Human Pirate {2}{U}, 2/3 (The Hobbit
// Eternal, HOC):
//
//	"{1}, {T}: Draw a card, then discard a card."
//
// Merfolk Looter's loot with a {1} in front of the tap (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1941f993-41d7-49e3-8891-41b3cecd5a9d",
		Name:         "Captain of Umbar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
		}},
	})
}
