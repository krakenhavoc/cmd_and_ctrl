package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Chariot — Artifact — Vehicle {4}, 3/3 (Theros Beyond Death, #5514):
//
//	"First strike, trample, haste
//	 Crew 1 (Tap any number of creatures you control with total power
//	 1 or more: This Vehicle becomes an artifact creature until end
//	 of turn.)"
//
// Dragonfly Suit's shape: the keywords sit on the Vehicle and matter
// once Crew 1 has made it an artifact creature (CR 702.122).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "589ea35b-c64e-4cf6-81c5-806e9463e992",
		Name:            "Thundering Chariot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "trample", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Thundering Chariot"),
		}},
	})
}
