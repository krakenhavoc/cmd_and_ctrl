package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consulate Dreadnought — Artifact — Vehicle {1}, 7/11 (Aether Revolt,
// #5536):
//
//	"Crew 6 (Tap any number of creatures you control with total power 6 or more: This Vehicle becomes an artifact creature until end of turn.)"
//
// A Vehicle with crew and nothing else: Colossal Plow's crew row
// (CR 702.122).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc822c9a-633f-4830-9cb3-9af6a80622e4",
		Name:         "Consulate Dreadnought",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 6",
			Cost:    effects.CrewCost(6),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Consulate Dreadnought"),
		}},
	})
}
