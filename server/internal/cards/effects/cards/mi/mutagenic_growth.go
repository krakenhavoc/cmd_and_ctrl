package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mutagenic Growth — Instant {G/P} (New Phyrexia, #5575):
//
//	"({G/P} can be paid with either {G} or 2 life.)
//	 Target creature gets +2/+2 until end of turn."
//
// Elvish Fury's pump, without the buyback. The Phyrexian symbol is paid
// by the engine at cast (CR 107.4f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c17b9e5b-58b6-4ccf-8283-945ba77429e7",
		Name:         "Mutagenic Growth",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(2, 2, "Mutagenic Growth — +2/+2"),
	})
}
