package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incendiary Flow — Sorcery {1}{R}:
//
//	Incendiary Flow deals 3 damage to any target. If a creature dealt damage this way would die this turn, exile it instead.
//
// "A creature dealt damage this way" is registered from the damage's continuation, on a creature that was dealt more than 0 damage (ADR 0108 §1 decision 3).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91b2ffe8-155d-4b9f-82dd-868cc895856b",
		Name:         "Incendiary Flow",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageAnyTargetExileIfDealtDies(effects.FixedAmount(3), false),
	})
}
