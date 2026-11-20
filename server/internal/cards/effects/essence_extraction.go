package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Essence Extraction — Instant {1}{B}{B} (Kaladesh, #5539):
//
//	"Essence Extraction deals 3 damage to target creature and you gain
//	 3 life."
//
// Vicious Hunger's body at 3: the damage from the spell, then the life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14363df6-ac0f-476a-bbd8-ce351de0babf",
		Name:         "Essence Extraction",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), LifeGain: 3, Answers: game.AnswerRemove},
		OnResolve:    ResolveSteps(damageTheClauseTarget(0, 3), GainLife{Amount: 3}),
	})
}
