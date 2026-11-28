package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magma Mine — Artifact {1} (Visions, #5631):
//
//	"{4}: Put a pressure counter on this artifact.
//	 {T}, Sacrifice this artifact: It deals damage equal to the number of
//	 pressure counters on it to any target."
//
// The second ability's damage is the Mine's as it last existed (CR 608.2h):
// its pressure counters at the moment it was sacrificed to pay the cost,
// dealt by it.
//
// No simplification.
func init() {
	const pressure = "pressure"
	effects.Register(effects.Spec{
		OracleID:     "8a2726a4-29b7-4836-bd0a-9c0d914ed915",
		Name:         "Magma Mine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}: Put a pressure counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{4}"),
				Effect:  effects.PutACounterOnThis(pressure),
			},
			{
				Label:   "{T}, Sacrifice this artifact: It deals damage equal to the number of pressure counters on it to any target.",
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Effect:  effects.SacrificedSourceCounterDamage(pressure),
			},
		},
	})
}
