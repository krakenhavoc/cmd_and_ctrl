package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "30399525-ab58-4ed0-8d87-3decb0dd4862",
		Name:         "Research Assistant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
