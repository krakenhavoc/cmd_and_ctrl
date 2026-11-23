package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Masked Meower — Creature — Spider Cat Hero {R}, 1/1 (Marvel Super Heroes
// Commander, #4455):
//
//	"Haste
//	 Discard a card, Sacrifice this creature: Draw a card."
//
// Haste rides PrintedKeywords. Both costs are paid together at announce
// (CR 602.2b): the discard is picked from hand, and the sacrifice is the
// Meower itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a0037ff8-64db-45dd-bba1-db5c49095083",
		Name:            "Masked Meower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.DiscardACard(), effects.SacrificeThis()),
			Effect:  effects.EbDrawACard,
		}},
	})
}
