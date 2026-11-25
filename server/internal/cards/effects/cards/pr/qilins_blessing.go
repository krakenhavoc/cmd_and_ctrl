package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Qilin's Blessing — Instant {W} (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"Target creature gets +2/+2 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31585036-cdc3-47f5-9fc7-913212826f2a",
		Name:         "Qilin's Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(2, 2, "Qilin's Blessing — +2/+2 until end of turn"),
	})
}
