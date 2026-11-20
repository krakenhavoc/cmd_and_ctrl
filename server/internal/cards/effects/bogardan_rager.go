package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "abbbfc22-18cc-4799-b02c-190173936872",
		Name:            "Bogardan Rager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters(label, boostTargetUntilEOT(label, 4, 0)),
				TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 4, 0))),
		},
	})
}
