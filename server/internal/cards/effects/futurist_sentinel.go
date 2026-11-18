package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Futurist Sentinel — Artifact — Vehicle {3}{U}, 6/6 (Neon Dynasty):
//
//	"Crew 3"
//
// The Vehicle is not a creature until a crew ability resolves
// (CR 702.122), so any printed keyword sits inert on the artifact until then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78d9faff-2ad9-4cf9-a239-27c50435d007",
		Name:         "Futurist Sentinel",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Futurist Sentinel"),
		}},
	})
}
