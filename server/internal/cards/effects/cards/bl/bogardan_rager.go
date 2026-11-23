package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bogardan Rager — Creature — Elemental {5}{R}, 3/4 (Time Spiral, #5596):
//
//	"Flash
//	 When this creature enters, target creature gets +4/+0 until end of
//	 turn."
//
// Flash is printed; the trigger targets as it is put on the stack and
// pumps the target if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Bogardan Rager — target creature gets +4/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "abbbfc22-18cc-4799-b02c-190173936872",
		Name:            "Bogardan Rager",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters(label, effects.BoostTargetUntilEOT(label, 4, 0)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 4, 0))),
		},
	})
}
