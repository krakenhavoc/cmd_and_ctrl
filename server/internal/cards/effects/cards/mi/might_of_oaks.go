package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Might of Oaks — Instant {3}{G} (Urza's Legacy, #5623):
//
//	"Target creature gets +7/+7 until end of turn."
//
// Giant Growth at +7/+7, on the shared boostTheTargetUntilEOT body. The
// bonus ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8331f281-819b-4a0b-bad7-bd86dbedb877",
		Name:         "Might of Oaks",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 7, 7)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(7, 7, "Might of Oaks — +7/+7"),
	})
}
