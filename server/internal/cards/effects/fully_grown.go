package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fully Grown — Instant {2}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gets +3/+3 until end of turn. Put a trample counter on it."
//
// The pump ends at cleanup; the trample counter stays on the creature
// (CR 122.1b). Each step skips a target that is gone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c98163f4-4a78-43f8-9855-4a23ebb9ceeb",
		Name:         "Fully Grown",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    ResolveSteps(ikoBoostTheTargetUntilEOT(3, 3, "Fully Grown — +3/+3"), ikoCounterOnTheTarget(game.CounterTrample)),
	})
}
