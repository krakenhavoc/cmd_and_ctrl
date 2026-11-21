package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ripclaw Wrangler — Artifact — Vehicle {3}{B}, 4/3 (Aetherdrift, #4464):
//
//	"When this Vehicle enters, each opponent discards a card.
//	 Crew 2 (Tap any number of creatures you control with total power 2
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// The enters trigger is Burglar Rat's (eachOpponentDiscardsOne): one
// discard prompt per opponent, each picking from their own hand
// (CR 701.9a). It triggers whether or not the Vehicle is a creature.
// Crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b2ffdb8-3681-4d64-ac03-8c860082e60e",
		Name:         "Ripclaw Wrangler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ripclaw Wrangler — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Ripclaw Wrangler"),
		}},
	})
}
