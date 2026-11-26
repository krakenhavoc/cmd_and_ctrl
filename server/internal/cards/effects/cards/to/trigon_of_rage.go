package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trigon of Rage — Artifact {2} (Scars of Mirrodin, S82 OLD batch 470, #5577):
//
//	"This artifact enters with three charge counters on it.
//	 {R}{R}, {T}: Put a charge counter on this artifact.
//	 {2}, {T}, Remove a charge counter from this artifact: Target creature gets
//	 +3/+0 until end of turn."
//
// The Trigon shape (OldB470Trigon) with a targeted until-end-of-turn pump.
//
// No simplification.
func init() {
	effects.Register(effects.OldB470Trigon(
		effects.Spec{OracleID: "3b172be8-1bac-441e-98cc-f5af8da6005b", Name: "Trigon of Rage", Completeness: effects.CompletenessFull},
		"{R}{R}", "Target creature gets +3/+0 until end of turn.",
		effects.TargetCreature("target creature"), effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0)), game.AnswerPump),
		effects.BoostTargetUntilEOT("Trigon of Rage — +3/+0 until end of turn", 3, 0)))
}
