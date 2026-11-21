package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sorin's Thirst — Instant {B}{B} (Magic 2012, #5573):
//
//	"Sorin's Thirst deals 2 damage to target creature and you gain 2
//	 life."
//
// Vicious Hunger's text: 2 damage to the target creature, then the life
// gain, which waits for any prompt the damage asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff27ff37-96c0-41af-8881-a078e884e67b",
		Name:         "Sorin's Thirst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), LifeGain: 2, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 2), effects.GainLife{Amount: 2}),
	})
}
