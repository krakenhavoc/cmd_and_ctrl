package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devotee of Strength — Creature — Snake Wizard {2}{G}, 3/2 (Hour of
// Devastation, #5534):
//
//	"{4}{G}: Target creature gets +2/+2 until end of turn."
//
// Deepwood Drummer's row without its tap and discard: the shared
// boostTargetUntilEOT body at +2/+2, declared a benefit to the target. A
// target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6d8f944-f516-49d7-9aa8-77b98e862d6a",
		Name:         "Devotee of Strength",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{4}{G}: Target creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump,
				Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2)),
			},
			Cost:    ManaCost("{4}{G}"),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Devotee of Strength — +2/+2", 2, 2),
		}},
	})
}
