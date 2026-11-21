package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Debris Beetle — Artifact — Vehicle {2}{B}{G}, 6/6 (Aetherdrift, #4464):
//
//	"Trample
//	 When this Vehicle enters, each opponent loses 3 life and you gain 3
//	 life.
//	 Crew 2"
//
// Acolyte of Aclazotz's drain as an enters trigger, for three: each
// opponent loses the life (no damage, so nothing prevents it) and you gain
// a fixed 3 however much they lost. The trigger watches the Vehicle
// entering, not becoming a creature, so it needs no crew. Trample is the
// Vehicle's whether or not it is crewed; crew is the engine's
// (CR 702.122a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b15d571-2f90-44f7-9153-20ed51253045",
		Name:            "Debris Beetle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Debris Beetle — each opponent loses 3 life and you gain 3 life",
				effects.Do(effects.EachOpponentLosesLife{N: 3}, effects.GainLife{Amount: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Debris Beetle"),
		}},
	})
}
