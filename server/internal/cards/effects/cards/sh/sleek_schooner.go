package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sleek Schooner — Artifact — Vehicle {3}, 4/3 (Ixalan, S82 older sets):
//
//	"Crew 1"
//
// The plain crew ability (Aethersphere Harvester's), crewed by tapping any
// creatures with total power 1 or more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3284010f-aeb2-493e-9252-3b953a192a94",
		Name:         "Sleek Schooner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Sleek Schooner"),
		}},
	})
}
