package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howling Fury — Sorcery {2}{B} (Portal, #5630):
//
//	"Target creature gets +4/+0 until end of turn."
//
// The shared "target creature gets +P/+T until end of turn" body
// (boostTheTargetUntilEOT). The pump ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eda60752-d225-4fd0-9f0f-9b99e321b8fa",
		Name:         "Howling Fury",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(4, 0, "Howling Fury — +4/+0"),
	})
}
