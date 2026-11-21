package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giantfall — Instant {1}{R} (Lorwyn Eclipsed, #4458):
//
//	"Choose one —
//	 • Target creature you control deals damage equal to its power to
//	   target creature an opponent controls.
//	 • Destroy target artifact."
//
// HULK SMASH!'s two bullets without its teamwork: the bite is two clauses
// in one mode, read per slot, and the damage is the biter's power as the
// spell resolves; nothing is dealt if either creature has become an
// illegal target (CR 608.2b). The second bullet destroys any artifact,
// creature or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77e5e7e3-b97b-4953-9df4-1aee509fbf2a",
		Name:         "Giantfall",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target creature you control deals damage equal to its power to target creature an opponent controls.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				),
				effects.BiteTheModesSecondTarget),
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
		),
	})
}
