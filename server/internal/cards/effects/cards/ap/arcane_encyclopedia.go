package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcane Encyclopedia — Artifact — Book {3} (Core Set 2019, #5527):
//
//	"{3}, {T}: Draw a card."
//
// A mana-and-tap ability that draws one card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d91ea728-2d69-42cb-bb5d-e2b058f0d7b1",
		Name:         "Arcane Encyclopedia",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
