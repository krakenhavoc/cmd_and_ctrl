package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soothsayer Adept — Creature — Merfolk Wizard {1}{U}, 1/3 (Strixhaven:
// School of Mages, #5502):
//
//	"{1}{U}, {T}: Draw a card, then discard a card."
//
// Strix Lookout's loot row: the draw, then the discard prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29723d39-aaca-48fd-ba4a-14931e4483e9",
		Name:         "Soothsayer Adept",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
