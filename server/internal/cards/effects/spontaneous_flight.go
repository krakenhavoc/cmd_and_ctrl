package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spontaneous Flight — Instant {2}{W} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gets +2/+2 until end of turn. Put a flying counter on it."
//
// The pump ends at cleanup; the flying counter stays on the creature
// (CR 122.1b). Each step skips a target that is gone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04da8c46-9703-47f2-974b-c144464ec380",
		Name:         "Spontaneous Flight",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    ResolveSteps(ikoBoostTheTargetUntilEOT(2, 2, "Spontaneous Flight — +2/+2"), ikoCounterOnTheTarget(game.CounterFlying)),
	})
}
