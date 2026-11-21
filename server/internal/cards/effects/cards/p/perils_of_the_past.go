package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f1f77fb4-2c83-4cbf-894e-624abe6cd95f",
		Name:         "Perils of the Past",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Use the Relic — Target creature gets +3/+2 until end of turn.",
				effects.TargetCreature("target creature"), effects.BoostTheModesTarget(3, 2, "Perils of the Past")), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 2))),
			effects.ModeDoing("Destroy the Relic — Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()), effects.DestroyTheModesTarget),
		),
	})
}
