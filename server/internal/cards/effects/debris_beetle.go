package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2b15d571-2f90-44f7-9153-20ed51253045",
		Name:            "Debris Beetle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Debris Beetle — each opponent loses 3 life and you gain 3 life",
				Do(EachOpponentLosesLife{N: 3}, GainLife{Amount: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Debris Beetle"),
		}},
	})
}
