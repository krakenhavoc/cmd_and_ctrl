package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haazda Officer — Creature — Human Soldier {2}{W}, 3/2 (Ravnica Allegiance,
// #5521):
//
//	"When this creature enters, target creature you control gets +1/+1
//	 until end of turn."
//
// Intrepid Rabbit's enters trigger: it targets as it goes on the stack (CR
// 603.3d) and does nothing to a target that has become illegal (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35023c04-ae62-47bf-80a0-faa2d7a67788",
		Name:         "Haazda Officer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Haazda Officer — target creature you control gets +1/+1 until end of turn",
				effects.BoostTargetUntilEOT("Haazda Officer — +1/+1", 1, 1)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1))),
		},
	})
}
