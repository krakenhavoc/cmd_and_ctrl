package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "345de988-4d26-4400-8202-225968db370f",
		Name:            "Ovalchase Dragster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Ovalchase Dragster"),
		}},
	})
}
