package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Shred — Sorcery {3}{B}{B} (Portal, S82 OLD batch 782, #5630):
//
//	"Soul Shred deals 3 damage to target nonblack creature. You gain 3 life."
//
// Morbid Hunger's two sentences as steps over a nonblack creature target: the
// life is gained whether or not the damage was prevented (CR 608.2c), but not
// at all when the target is illegal on resolution and the spell is countered
// by the rules (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2894ef5e-738b-4ece-b143-6662d9453295",
		Name:         "Soul Shred",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		Purpose:      game.Purpose{Targets: effects.ForTargets(effects.DamageToTarget(0, 3)).Targets, LifeGain: 3, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 3), effects.GainLife{Amount: 3}),
	})
}
