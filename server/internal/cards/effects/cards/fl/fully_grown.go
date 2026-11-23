package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fully Grown — Instant {2}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gets +3/+3 until end of turn. Put a trample counter on it."
//
// The pump ends at cleanup; the trample counter stays on the creature
// (CR 122.1b). Each step skips a target that is gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c98163f4-4a78-43f8-9855-4a23ebb9ceeb",
		Name:         "Fully Grown",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ResolveSteps(effects.IkoBoostTheTargetUntilEOT(3, 3, "Fully Grown — +3/+3"), effects.IkoCounterOnTheTarget(game.CounterTrample)),
	})
}
