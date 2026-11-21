package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elspeth's Smite — Instant {W}:
//
//	Elspeth's Smite deals 3 damage to target attacking or blocking creature. If that creature would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f404fe4-4335-4dcc-ba90-78246c4b880b",
		Name:         "Elspeth's Smite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageFirstTargetExileIfItDies(effects.FixedAmount(3)),
	})
}
