package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Patagia Tiger — Creature — Cat {4}{W}, 3/4 (Ikoria: Lair of Behemoths,
// S82 OLD batch 120, #5512):
//
//	"Flying
//	 When this creature enters, target Human you control gets +2/+2 until end
//	 of turn."
//
// Flying is a printed keyword. The enters trigger targets as it goes on the
// stack (CR 603.3d); "Human" is read through the layered view, so a
// changeling counts, and the Tiger is not one. A target that is gone or no
// longer a Human you control by resolution gets nothing (CR 608.2b). The
// +2/+2 is a layer 7c change that ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "256ab710-4210-4142-973f-22ca029d5474",
		Name:            "Patagia Tiger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Patagia Tiger — target Human you control gets +2/+2 until end of turn",
				effects.BoostTargetUntilEOT("Patagia Tiger — +2/+2", 2, 2)),
				effects.TargetCreature("target Human you control", effects.YouControl(), effects.HasSubtype("Human"))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))),
		},
	})
}
