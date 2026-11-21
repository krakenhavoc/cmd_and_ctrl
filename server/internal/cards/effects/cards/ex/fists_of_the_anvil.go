package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fists of the Anvil — Instant {1}{R} (Mirrodin, #5606):
//
//	"Target creature gets +4/+0 until end of turn."
//
// A layer 7c modification (CR 613.4c) pinned to the one target and
// swept at cleanup (CR 514.2); a target that is no longer legal leaves
// the spell with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc05e582-e760-4fe2-ba43-e9d8e63f3f85",
		Name:         "Fists of the Anvil",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(4, 0, "Fists of the Anvil — +4/+0 until end of turn"),
	})
}
