package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aradara Express — Artifact — Vehicle {5}, 8/6 (Kaladesh, #5539):
//
//	"Menace
//	 Crew 4 (Tap any number of creatures you control with total power 4
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// Air Response Unit's crew (vehicles.go, CR 702.122); the printed keywords
// matter once it is crewed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be96b9e1-10ef-4955-8d23-c39384def86d",
		Name:            "Aradara Express",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 4",
			Cost:    CrewCost(4),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Aradara Express"),
		}},
	})
}
