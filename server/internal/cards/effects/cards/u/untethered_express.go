package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Untethered Express — Artifact — Vehicle {4}, 4/4 (Aether Revolt,
// #5536):
//
//	"Trample
//	 Whenever this Vehicle attacks, put a +1/+1 counter on it.
//	 Crew 1 (Tap any number of creatures you control with total power 1 or more: This Vehicle becomes an artifact creature until end of turn.)"
//
// Predator Ooze's attack counter and the crew row. The counter stays
// after the Vehicle stops being a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "652f049d-770e-4bd7-b65f-fe56e618366a",
		Name:            "Untethered Express",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Untethered Express — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Untethered Express"),
		}},
	})
}
