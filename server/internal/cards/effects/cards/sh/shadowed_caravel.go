package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadowed Caravel — Artifact — Vehicle {2}, 2/2 (Ixalan, #5532):
//
//	"Whenever a creature you control explores, put a +1/+1 counter on this
//	 Vehicle.
//	 Crew 2"
//
// Merfolk Cave-Diver's explore trigger (CR 701.44b) with a counter on the
// Vehicle, which gets it whether or not it is crewed: the Caravel is a
// permanent, and the counter stays when it stops being a creature. Crew is
// the ordinary printed one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7986fcb6-bafb-4bbc-9f2f-fd9e2a5fcad1",
		Name:         "Shadowed Caravel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventExplored, effects.ACreatureYouControlExplored,
				"Shadowed Caravel — put a +1/+1 counter on this Vehicle",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Shadowed Caravel"),
		}},
	})
}
