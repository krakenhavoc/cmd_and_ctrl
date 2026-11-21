package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Futurist Sentinel — Artifact — Vehicle {3}{U}, 6/6 (Neon Dynasty):
//
//	"Crew 3"
//
// The Vehicle is not a creature until a crew ability resolves
// (CR 702.122), so any printed keyword sits inert on the artifact until then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78d9faff-2ad9-4cf9-a239-27c50435d007",
		Name:         "Futurist Sentinel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Futurist Sentinel"),
		}},
	})
}
