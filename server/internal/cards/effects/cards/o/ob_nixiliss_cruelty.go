package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ob Nixilis's Cruelty — Instant {2}{B}:
//
//	Target creature gets -5/-5 until end of turn. If that creature would die this turn, exile it instead.
//
// The -N/-N is applied first, then the replacement is registered on the target (ADR 0108 §1), so a creature the shrink kills is exiled.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c638957f-88bf-40c2-834c-2be39d73bf41",
		Name:         "Ob Nixilis's Cruelty",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -5, -5)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ShrinkFirstTargetExileIfItDies(5),
	})
}
