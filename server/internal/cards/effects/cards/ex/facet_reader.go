package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Facet Reader — Creature — Human Wizard {1}{U}, 1/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"{1}, {T}: Draw a card, then discard a card."
//
// Captain of Umbar's loot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7aefe7a9-d696-4e2e-b1a5-d041d9e4e213",
		Name:         "Facet Reader",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) },
		}},
	})
}
