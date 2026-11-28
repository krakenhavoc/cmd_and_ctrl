package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fool's Tome — Artifact — Book {4} (Tempest, S82 OLD batch 766, #5628):
//
//	"{2}, {T}: Draw a card. Activate only if you have no cards in hand."
//
// Sea Gate Wreckage's shape: the hand is checked as the ability is activated
// (CR 602.5b), and paying the cost does not change it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65589538-d3c1-4dfd-a328-71f57f556212",
		Name:         "Fool's Tome",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}, {T}: Draw a card. Activate only if you have no cards in hand.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Condition: effects.NoCardsInHand(),
			Effect:    effects.B36DrawOne,
		}},
	})
}
