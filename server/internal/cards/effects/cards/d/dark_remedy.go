package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Remedy — Instant {1}{B} (Core Set 2020, #5518):
//
//	"Target creature gets +1/+3 until end of turn."
//
// Artful Maneuver's pump body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7cc744b-2fbd-40e6-8a59-392535c17454",
		Name:         "Dark Remedy",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(1, 3, "Dark Remedy — +1/+3"),
	})
}
