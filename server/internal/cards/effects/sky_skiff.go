package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sky Skiff — Artifact — Vehicle {2}, 2/3 (Kaladesh, #5539):
//
//	"Flying
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
		OracleID:        "f2c43b85-8c42-4b23-84c2-bec8d6230039",
		Name:            "Sky Skiff",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Sky Skiff"),
		}},
	})
}
