package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magma Rift — Sorcery {2}{R} (Zendikar, S82 OLD batch 492, #5583):
//
//	"As an additional cost to cast this spell, sacrifice a land.
//	 Magma Rift deals 5 damage to target creature."
//
// The additional cost is the sacrifice of a land (SacrificeCost, as Shard
// Volley pays it), chosen as the spell is cast; the damage is dealt to the
// target creature if it is still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0996d48a-f92a-4aaf-b8a3-f7ec8ff986ca",
		Name:           "Magma Rift",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a land", effects.Land()),
		Targets:        effects.TargetCreature("target creature"),
		Purpose:        game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve:      effects.DamageToFirstTargetStep(5),
	})
}
