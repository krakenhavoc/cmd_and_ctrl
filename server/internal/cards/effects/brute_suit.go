package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brute Suit — Artifact — Vehicle {3}, 4/3 (Neon Dynasty):
//
//	"Vigilance
//	 Crew 1"
//
// Vigilance is the printed keyword. The Vehicle is not a creature until a crew ability resolves
// (CR 702.122), so any printed keyword sits inert on the artifact until then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bc64e396-7fa1-4a4e-a528-352737714a9c",
		Name:            "Brute Suit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Brute Suit"),
		}},
	})
}
