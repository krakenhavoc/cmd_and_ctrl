package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pinpoint Avalanche — Instant {3}{R}{R}:
//
//	"Pinpoint Avalanche deals 4 damage to target creature. The damage
//	 can't be prevented."
//
// The rider is the spell's own (Spec.SpellDamageCantBePrevented, ADR
// 0107 §5).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "d619cce4-036a-4da2-ae64-f43fc095ebe9",
		Name:                       "Pinpoint Avalanche",
		Completeness:               effects.CompletenessFull,
		SpellDamageCantBePrevented: effects.Always(),
		Targets:                    effects.TargetCreature("target creature"),
		Purpose:                    effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:                  effects.DamageToFirstTargetStep(4),
	})
}
