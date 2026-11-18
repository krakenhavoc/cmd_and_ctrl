package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonfly Suit — Artifact — Vehicle {2}{W}, 3/2 (Neon Dynasty):
//
//	"Flying
//	 Crew 1"
//
// Flying is the printed keyword. The Vehicle is not a creature until a crew ability resolves
// (CR 702.122), so any printed keyword sits inert on the artifact until then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f242d66c-011c-4b5b-a7d6-5e0bce257f73",
		Name:            "Dragonfly Suit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Dragonfly Suit"),
		}},
	})
}
