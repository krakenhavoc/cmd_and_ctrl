package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pack's Favor — Instant {2}{G} (Guilds of Ravnica, #5525):
//
//	"Convoke
//	 Target creature gets +3/+3 until end of turn."
//
// Convoke is the shared tap cost (CR 702.51a); the pump is Giant Growth's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "072288da-5ba0-4176-b0e6-b3e43ecab2ec",
		Name:         "Pack's Favor",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(3, 3, "Pack's Favor — +3/+3 until end of turn"),
	})
}
