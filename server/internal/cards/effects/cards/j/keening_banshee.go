package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keening Banshee — Creature — Spirit {2}{B}{B}, 2/2 (Ravnica: City of Guilds,
// S82 OLD batch 576, #5600):
//
//	"Flying
//	 When this creature enters, target creature gets -2/-2 until end of turn."
//
// Farbog Boneflinger's trigger: the creature is chosen as the trigger goes on
// the stack and shrunk as it resolves if it is still legal (CR 608.2b). The
// Banshee may target itself or any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "56b69e46-a1e4-4545-99e2-ee4385b7e429",
		Name:            "Keening Banshee",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Keening Banshee — target creature gets -2/-2 until end of turn",
				effects.BoostTargetUntilEOT("Keening Banshee — -2/-2", -2, -2)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
	})
}
