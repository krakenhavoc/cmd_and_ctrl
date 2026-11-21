package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Friendly Ghost — Creature — Spirit {3}{W}, 2/4 (Duskmourn: House of
// Horror, #4466):
//
//	"Flying
//	 When this creature enters, target creature gets +2/+4 until end of
//	 turn."
//
// An enters trigger with one creature target, the Ghost itself among the
// legal picks; the boost is boostTargetUntilEOT, skipped when the target
// has left in response (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Friendly Ghost — target creature gets +2/+4 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "e6a93240-8578-479e-b5e7-73f51b141cc8",
		Name:            "Friendly Ghost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters(label, effects.BoostTargetUntilEOT(label, 2, 4)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 4))),
		},
	})
}
