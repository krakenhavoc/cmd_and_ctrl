package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vicious Hunger — Sorcery {B}{B}
// (Nemesis, #5618):
//
//	"Vicious Hunger deals 2 damage to target creature and you gain 2 life."
//
// Morbid Hunger's body at 2 without flashback: damageTheClauseTarget,
// then the life. If the creature is an illegal target as the spell
// resolves, the spell does nothing at all (CR 608.2b), so no life is
// gained.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38ea22cd-2c5d-4f66-a111-207aca4c67c3",
		Name:         "Vicious Hunger",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), LifeGain: 2, Answers: game.AnswerRemove},
		OnResolve:    ResolveSteps(damageTheClauseTarget(0, 2), GainLife{Amount: 2}),
	})
}
