package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gather Courage — Instant {G} (Ravnica: City of Guilds, #5600):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Target creature gets +2/+2 until end of turn."
//
// Convoke is the engine's tap cost (CR 702.51); the pump is the one target's
// +2/+2 until cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae033eb6-47ec-4b9e-9cdb-6d023aca07d1",
		Name:         "Gather Courage",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve:    effects.BoostTheTargetUntilEOT(2, 2, "Gather Courage — +2/+2"),
	})
}
