package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e23d88c4-8c9c-41cb-b4ff-ebb3f880ab63",
		Name:            "Viashino Bladescout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters(label, pumpAndGrantFirstTarget(0, 0, label, "first strike")),
				TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 0, 0, "first strike"))),
		},
	})
}
