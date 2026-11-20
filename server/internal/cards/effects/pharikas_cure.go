package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pharika's Cure — Instant {B}{B} (Theros, #5561):
//
//	"Pharika's Cure deals 2 damage to target creature and you gain 2 life."
//
// Vicious Hunger's body: 2 damage to the creature, then 2 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06d77149-52b6-4d2c-bfb9-1c25f6b1b0bb",
		Name:         "Pharika's Cure",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), LifeGain: 2, Answers: game.AnswerRemove},
		OnResolve:    ResolveSteps(damageTheClauseTarget(0, 2), GainLife{Amount: 2}),
	})
}
