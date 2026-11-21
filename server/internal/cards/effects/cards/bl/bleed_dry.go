package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bleed Dry — Instant {2}{B}{B}:
//
//	Target creature gets -13/-13 until end of turn. If that creature would die this turn, exile it instead.
//
// The -N/-N is applied first, then the replacement is registered on the target (ADR 0108 §1), so a creature the shrink kills is exiled.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c3faf4f-83c1-4098-98b8-bae15d59b0de",
		Name:         "Bleed Dry",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -13, -13)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ShrinkFirstTargetExileIfItDies(13),
	})
}
