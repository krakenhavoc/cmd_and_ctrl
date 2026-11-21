package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pharika's Cure — Instant {B}{B} (Theros, #5561):
//
//	"Pharika's Cure deals 2 damage to target creature and you gain 2 life."
//
// Vicious Hunger's body: 2 damage to the creature, then 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06d77149-52b6-4d2c-bfb9-1c25f6b1b0bb",
		Name:         "Pharika's Cure",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), LifeGain: 2, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 2), effects.GainLife{Amount: 2}),
	})
}
