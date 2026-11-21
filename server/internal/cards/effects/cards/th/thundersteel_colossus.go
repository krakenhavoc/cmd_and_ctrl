package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "9cef3013-6a84-4edd-9a47-e842df2200e3",
		Name:            "Thundersteel Colossus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Thundersteel Colossus"),
		}},
	})
}
