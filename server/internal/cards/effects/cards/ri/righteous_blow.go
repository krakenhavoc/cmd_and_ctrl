package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Righteous Blow — Instant {W} (Guilds of Ravnica, #5525):
//
//	"Righteous Blow deals 2 damage to target attacking or blocking creature."
//
// The target must be attacking or blocking as the spell is cast and again as
// it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc2bf46f-2c9b-4f85-b8d6-d13a3b64436d",
		Name:         "Righteous Blow",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
