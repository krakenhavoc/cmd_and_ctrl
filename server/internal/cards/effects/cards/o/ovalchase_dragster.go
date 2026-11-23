package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ovalchase Dragster — Artifact — Vehicle {4}, 6/1 (Kaladesh, #5539):
//
//	"Trample, haste
//	 Crew 1 (Tap any number of creatures you control with total power 1
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// Air Response Unit's crew (vehicles.go, CR 702.122); the printed keywords
// matter once it is crewed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "345de988-4d26-4400-8202-225968db370f",
		Name:            "Ovalchase Dragster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Ovalchase Dragster"),
		}},
	})
}
