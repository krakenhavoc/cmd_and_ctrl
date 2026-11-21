package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obliterating Bolt — Sorcery {1}{R}:
//
//	Obliterating Bolt deals 4 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57fe941c-a830-4570-afe2-18f93c7a7b84",
		Name:         "Obliterating Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:    effects.DamageFirstTargetExileIfItDies(effects.FixedAmount(4)),
	})
}
