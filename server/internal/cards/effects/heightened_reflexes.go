package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heightened Reflexes — Instant {R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gets +1/+0 until end of turn. Put a first strike counter on it."
//
// The pump ends at cleanup; the first strike counter stays on the creature
// (CR 122.1b). Each step skips a target that is gone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "513fd896-ab78-44fb-b136-f7c93288dc63",
		Name:         "Heightened Reflexes",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    ResolveSteps(ikoBoostTheTargetUntilEOT(1, 0, "Heightened Reflexes — +1/+0"), ikoCounterOnTheTarget(game.CounterFirstStrike)),
	})
}
