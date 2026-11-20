package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sudden Spinnerets — Instant {G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gets +1/+3 until end of turn. Put a reach counter on it. Untap it."
//
// The pump ends at cleanup; the reach counter stays on the creature
// (CR 122.1b). Each step skips a target that is gone (CR 608.2b), the
// untap last.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b83fede-851c-4e9d-98ba-d88f6caffc1b",
		Name:         "Sudden Spinnerets",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 3)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    ResolveSteps(ikoBoostTheTargetUntilEOT(1, 3, "Sudden Spinnerets — +1/+3"), ikoCounterOnTheTarget(game.CounterReach), ikoUntapTheTarget()),
	})
}
