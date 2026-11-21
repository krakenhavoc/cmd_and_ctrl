package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain of Umbar — Creature — Human Pirate {2}{U}, 2/3 (The Hobbit
// Eternal, HOC):
//
//	"{1}, {T}: Draw a card, then discard a card."
//
// Merfolk Looter's loot with a {1} in front of the tap (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1941f993-41d7-49e3-8891-41b3cecd5a9d",
		Name:         "Captain of Umbar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) },
		}},
	})
}
