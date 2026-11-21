package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phytoburst — Sorcery {1}{G}:
//
//	"Target creature gets +5/+5 until end of turn."
//
// The shared boostTheTargetUntilEOT body at +5/+5, ending at cleanup
// (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d68d175f-f59c-4995-84d4-fc593280194b",
		Name:         "Phytoburst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 5, 5)),
		OnResolve:    effects.BoostTheTargetUntilEOT(5, 5, "Phytoburst — +5/+5 until end of turn"),
	})
}
