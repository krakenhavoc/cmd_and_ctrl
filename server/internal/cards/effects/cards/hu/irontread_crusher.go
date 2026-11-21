package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Irontread Crusher — Artifact — Vehicle {4}, 6/6 (Aether Revolt,
// #5536):
//
//	"Crew 3 (Tap any number of creatures you control with total power 3 or more: This Vehicle becomes an artifact creature until end of turn.)"
//
// A Vehicle with crew and nothing else: Colossal Plow's crew row
// (CR 702.122).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c951fca7-f333-4ca3-b5d6-83967d1d093d",
		Name:         "Irontread Crusher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Irontread Crusher"),
		}},
	})
}
