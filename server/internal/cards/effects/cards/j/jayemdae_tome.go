package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jayemdae Tome — Artifact — Book {4} (Magic Origins):
//
//	"{4}, {T}: Draw a card."
//
// A mana-and-tap ability that draws one card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39ee576a-0803-4063-9c84-f2b537e4d44c",
		Name:         "Jayemdae Tome",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
