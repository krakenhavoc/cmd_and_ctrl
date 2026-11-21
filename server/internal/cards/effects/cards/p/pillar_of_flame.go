package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pillar of Flame — Sorcery {R}:
//
//	Pillar of Flame deals 2 damage to any target. If a creature dealt damage this way would die this turn, exile it instead.
//
// "A creature dealt damage this way" is registered from the damage's continuation, on a creature that was dealt more than 0 damage (ADR 0108 §1 decision 3).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "468cfc88-a493-44dc-9d0a-63d9cc89c114",
		Name:         "Pillar of Flame",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve:    effects.DamageAnyTargetExileIfDealtDies(effects.FixedAmount(2), false),
	})
}
