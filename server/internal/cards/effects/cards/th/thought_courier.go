package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thought Courier — Creature — Human Wizard {1}{U}, 1/1 (Fifth Dawn, #5604):
//
//	"{T}: Draw a card, then discard a card."
//
// Merfolk Looter's loot: draw first, then the discard prompt (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06ea02f5-f296-4334-bb6f-9cb018e1e150",
		Name:         "Thought Courier",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) },
		}},
	})
}
