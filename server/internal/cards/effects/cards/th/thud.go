package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thud — Sorcery {R}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Thud deals damage equal to the sacrificed creature's power to any
//	 target."
//
// Fling at sorcery speed. The damage is the sacrificed creature's power
// as it last existed on the battlefield (the 2018-07-13 ruling, CR
// 608.2h), read off the payment record (ADR 0113 §1); negative power
// deals none (CR 107.1b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "5d2a9859-1353-4431-9343-f5999450acd1",
		Name:           "Thud",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetAny(),
		OnResolve:      effects.DamageEqualToSacrificedPower,
	})
}
