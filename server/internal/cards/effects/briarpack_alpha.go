package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Briarpack Alpha — Creature — Wolf {3}{G}, 3/3 (Dark Ascension, #5571):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.)
//	 When this creature enters, target creature gets +2/+2 until end
//	 of turn."
//
// Daybreak Combatants' enters trigger with +2/+2; the target, the Alpha
// included, is picked as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b). Flash is the printed keyword the cast-timing
// gate reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5604c475-6e01-4e9d-b4e6-8e8d72874581",
		Name:            "Briarpack Alpha",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Briarpack Alpha — target creature gets +2/+2 until end of turn",
				boostTargetUntilEOT("Briarpack Alpha — +2/+2", 2, 2)),
				TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 2, 2))),
		},
	})
}
