package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Command the Storm — Instant {4}{R} (Guilds of Ravnica, #5525):
//
//	"Command the Storm deals 5 damage to target creature."
//
// Concentrated Fire's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1dc816b5-8918-4a0c-bb10-74abe579a98c",
		Name:         "Command the Storm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}
