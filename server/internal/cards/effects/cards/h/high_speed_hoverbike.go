package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High-Speed Hoverbike — Artifact — Vehicle {2}, 2/2 (Neon Dynasty):
//
//	"Flash
//	 Flying
//	 When this Vehicle enters, tap up to one target creature.
//	 Crew 1"
//
// Flash and flying are printed keywords. The entry trigger is Eddymurk
// Crab's tap over up to one creature (CR 601.2c: choosing none is legal);
// a target gone by resolution is skipped (CR 608.2b). The Vehicle is not a
// creature until crewed (CR 702.122).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("High-Speed Hoverbike — tap up to one target creature", effects.TapEachLegalTarget)
	enters.Targets = effects.TargetCreature("up to one target creature").WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:        "9bf2f0d9-57fb-4e44-90a5-260d735076a0",
		Name:            "High-Speed Hoverbike",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered:       []game.TriggeredAbility{enters},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("High-Speed Hoverbike"),
		}},
	})
}
