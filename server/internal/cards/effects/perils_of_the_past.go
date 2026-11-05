package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perils of the Past — Instant {1}{R} (Star Trek, #4973):
//
//	"Choose one —
//	 • Use the Relic — Target creature gets +3/+2 until end of turn.
//	 • Destroy the Relic — Destroy target artifact."
//
// "Use the Relic" and "Destroy the Relic" are mode names with no rules
// meaning of their own. The engine applies the chosen bullet's target
// clause, so the picker offers creatures or artifacts by mode.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1f77fb4-2c83-4cbf-894e-624abe6cd95f",
		Name:         "Perils of the Past",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Use the Relic — Target creature gets +3/+2 until end of turn.",
				TargetCreature("target creature"), BoostTheModesTarget(3, 2, "Perils of the Past")), ForTargets(BenefitUntilEOT(0, 3, 2))),
			ModeDoing("Destroy the Relic — Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()), DestroyTheModesTarget),
		),
	})
}
