package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heightened Reflexes — Instant {R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gets +1/+0 until end of turn. Put a first strike counter on it."
//
// The pump ends at cleanup; the first strike counter stays on the creature
// (CR 122.1b). Each step skips a target that is gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "513fd896-ab78-44fb-b136-f7c93288dc63",
		Name:         "Heightened Reflexes",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ResolveSteps(effects.IkoBoostTheTargetUntilEOT(1, 0, "Heightened Reflexes — +1/+0"), effects.IkoCounterOnTheTarget(game.CounterFirstStrike)),
	})
}
