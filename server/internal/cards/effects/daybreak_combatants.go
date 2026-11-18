package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daybreak Combatants — Creature — Human Warrior {2}{R}, 2/2 (Innistrad:
// Crimson Vow):
//
//	"Haste
//	 When this creature enters, target creature gets +2/+0 until end of
//	 turn."
//
// The target may be the Combatants itself or any other creature; it is
// chosen as the trigger goes on the stack and re-checked as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fbb3f959-ded7-4570-bacf-0e1394dbef80",
		Name:            "Daybreak Combatants",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Daybreak Combatants — target creature gets +2/+0",
				boostTargetUntilEOT("Daybreak Combatants — +2/+0", 2, 0)),
				TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 2, 0))),
		},
	})
}
