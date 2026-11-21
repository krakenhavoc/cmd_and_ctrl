package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strix Lookout — Creature — Bird {1}{U}, 1/2 (Foundations, #4465):
//
//	"Flying, vigilance
//	 {1}{U}, {T}: Draw a card, then discard a card."
//
// Flying Drone's looter without the discount. The discard is asked
// after the draw, so the drawn card may be the one discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5a6d01c8-47b6-43e5-a6e9-c48088286007",
		Name:            "Strix Lookout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
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
