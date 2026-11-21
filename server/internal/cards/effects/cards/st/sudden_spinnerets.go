package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5b83fede-851c-4e9d-98ba-d88f6caffc1b",
		Name:         "Sudden Spinnerets",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ResolveSteps(effects.IkoBoostTheTargetUntilEOT(1, 3, "Sudden Spinnerets — +1/+3"), effects.IkoCounterOnTheTarget(game.CounterReach), effects.IkoUntapTheTarget()),
	})
}
