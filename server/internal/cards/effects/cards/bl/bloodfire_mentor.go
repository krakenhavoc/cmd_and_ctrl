package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodfire Mentor — Creature — Efreet Shaman {2}{R} 0/5 (Khans of Tarkir):
//
//	"{2}{U}, {T}: Draw a card, then discard a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c47319a-4062-4aec-b68a-e8e8f023cf86",
		Name:         "Bloodfire Mentor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
