package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yamabushi's Flame — Instant {2}{R}:
//
//	Yamabushi's Flame deals 3 damage to any target. If a creature dealt damage this way would die this turn, exile it instead.
//
// "A creature dealt damage this way" is registered from the damage's continuation, on a creature that was dealt more than 0 damage (ADR 0108 §1 decision 3).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0462e985-c99e-4404-b212-e9d8baecce72",
		Name:         "Yamabushi's Flame",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageAnyTargetExileIfDealtDies(effects.FixedAmount(3), false),
	})
}
