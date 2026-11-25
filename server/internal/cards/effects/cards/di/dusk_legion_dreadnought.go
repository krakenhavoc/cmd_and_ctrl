package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dusk Legion Dreadnought — Artifact — Vehicle {5}, 4/6 (Ixalan, S82 older
// sets):
//
//	"Vigilance
//	 Crew 2"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7278f382-8249-4981-9d87-47cef7dc13dd",
		Name:            "Dusk Legion Dreadnought",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Dusk Legion Dreadnought"),
		}},
	})
}
