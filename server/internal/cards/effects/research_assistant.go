package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Research Assistant — Creature — Human Wizard {1}{U}, 1/3 (Magic 2015,
// #5556):
//
//	"{3}{U}, {T}: Draw a card, then discard a card."
//
// Frontline Sage's loot at {3}{U}: the discard opens only once the draw
// has resolved (lootOne). The {T} means a summoning-sick Assistant can't
// activate it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30399525-ab58-4ed0-8d87-3decb0dd4862",
		Name:         "Research Assistant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Cost:    Plus(ManaCost("{3}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
	})
}
