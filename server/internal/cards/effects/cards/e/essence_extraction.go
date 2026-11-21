package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Extraction — Instant {1}{B}{B} (Kaladesh, #5539):
//
//	"Essence Extraction deals 3 damage to target creature and you gain
//	 3 life."
//
// Vicious Hunger's body at 3: the damage from the spell, then the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14363df6-ac0f-476a-bbd8-ce351de0babf",
		Name:         "Essence Extraction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), LifeGain: 3, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 3), effects.GainLife{Amount: 3}),
	})
}
