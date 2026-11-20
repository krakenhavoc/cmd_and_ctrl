package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Essence Drain — Sorcery {4}{B} (Darksteel, #5605):
//
//	"Essence Drain deals 3 damage to any target and you gain 3 life."
//
// Morbid Hunger's two steps without its flashback: the life gain waits
// for any prompt the damage asks (CR 608.2c). With its only target gone
// by resolution the spell does nothing, the life gain included (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bf5402f-7bba-4c0c-aebd-1b231768f2b2",
		Name:         "Essence Drain",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), LifeGain: 3, Answers: game.AnswerRemove},
		OnResolve:    ResolveSteps(damageTheClauseTarget(0, 3), GainLife{Amount: 3}),
	})
}
