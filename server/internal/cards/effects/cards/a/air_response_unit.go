package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Air Response Unit — Artifact — Vehicle {2}{W}, 3/3 (Aetherdrift,
// #4464):
//
//	"Flying, vigilance
//	 Crew 1"
//
// Two printed keywords and the engine's crew (CR 702.122a). The keywords
// are the Vehicle's whether or not it is crewed, and matter once it is a
// creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "18a22d9f-81bc-4ed5-bd83-e3f8ec5e38c8",
		Name:            "Air Response Unit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Air Response Unit"),
		}},
	})
}
