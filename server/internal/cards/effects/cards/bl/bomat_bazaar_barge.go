package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bomat Bazaar Barge — Artifact — Vehicle {4}, 5/5 (Kaladesh, #5539):
//
//	"When this Vehicle enters, draw a card.
//	 Crew 3"
//
// Ribskiff's shape: a plain entry draw and the ordinary crew ability, with no
// other text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c46a7dc0-88d7-4fbe-8d63-a40688bb5861",
		Name:         "Bomat Bazaar Barge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bomat Bazaar Barge — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Bomat Bazaar Barge"),
		}},
	})
}
