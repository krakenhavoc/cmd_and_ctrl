package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carbonize — Instant {2}{R}:
//
//	Carbonize deals 3 damage to any target. If it's a creature, it can't be regenerated this turn, and if it would die this turn, exile it instead.
//
// Both riders are the spell's, not the damage's (the Disintegrate shape): a creature target is marked whether or not the damage is dealt (ADR 0108 §1, §2).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a32795a2-a965-4a85-9944-fd9eed464e65",
		Name:         "Carbonize",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageThenIfCreatureNoRegenExileIfDies(effects.FixedAmount(3)),
	})
}
