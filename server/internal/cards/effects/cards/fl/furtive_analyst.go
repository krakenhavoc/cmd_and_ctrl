package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furtive Analyst — Creature — Human Wizard 1/4 {2}{U} (March of the Machine, S82):
//
//	"Vigilance
//	 {2}, {T}: Draw a card, then discard a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e5c6bbee-5e6c-4a59-bc5d-285c5f5cf228",
		Name:            "Furtive Analyst",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
