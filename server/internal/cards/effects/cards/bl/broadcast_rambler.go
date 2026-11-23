package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broadcast Rambler — Artifact — Vehicle {4}{W}, 5/4 (Aetherdrift,
// #4464):
//
//	"When this Vehicle enters, create a 1/1 colorless Thopter artifact
//	 creature token with flying.
//	 Crew 1"
//
// Turtle Blimp's shape: an enters trigger that makes a token, and the
// engine's crew (CR 702.122a). The Thopter can crew the Rambler the turn
// it is made: crewing taps, and summoning sickness does not stop a
// creature from paying a crew cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21aaecc7-b9ec-4068-8e47-4dffaa12f102",
		Name:         "Broadcast Rambler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Broadcast Rambler — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Broadcast Rambler"),
		}},
	})
}
