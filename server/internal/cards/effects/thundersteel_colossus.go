package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thundersteel Colossus — Artifact — Vehicle {7}, 7/7 (Neon Dynasty):
//
//	"Trample, haste
//	 Crew 2"
//
// Trample and haste are the printed keywords. The Vehicle is not a creature until a crew ability resolves
// (CR 702.122), so any printed keyword sits inert on the artifact until then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9cef3013-6a84-4edd-9a47-e842df2200e3",
		Name:            "Thundersteel Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Thundersteel Colossus"),
		}},
	})
}
