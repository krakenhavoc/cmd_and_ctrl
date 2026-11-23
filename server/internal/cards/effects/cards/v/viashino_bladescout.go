package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Bladescout — Creature — Lizard Scout {1}{R}, 2/1 (Time
// Spiral, #5596):
//
//	"Flash
//	 When this creature enters, target creature gains first strike until
//	 end of turn."
//
// Flash is printed; the trigger is Icatian Scout's grant as an enters
// trigger (pumpAndGrantFirstTarget with no pump).
//
// No simplification.
func init() {
	const label = "Viashino Bladescout — target creature gains first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "e23d88c4-8c9c-41cb-b4ff-ebb3f880ab63",
		Name:            "Viashino Bladescout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters(label, effects.PumpAndGrantFirstTarget(0, 0, label, "first strike")),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike"))),
		},
	})
}
