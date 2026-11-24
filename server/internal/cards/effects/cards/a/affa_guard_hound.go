package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Affa Guard Hound — Creature — Dog {2}{W}, 2/2 (Jumpstart, #5510):
//
//	"Flash
//	 When this creature enters, target creature gets +0/+3 until end of
//	 turn."
//
// Flash is a printed keyword. The target is chosen as the trigger goes on
// the stack and may be the Hound itself; the boost is a layer 7c +0/+3
// until end of turn, skipped for a target that is no longer legal.
//
// No simplification.
func init() {
	const label = "Affa Guard Hound — target creature gets +0/+3 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "3819daab-fe35-4f99-9282-47bd05c7fcd6",
		Name:            "Affa Guard Hound",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters(label, effects.S82OldB098TargetCreatureGetsUntilEOT(label, 0, 3)),
				effects.TargetCreature("target creature")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 3))),
		},
	})
}
