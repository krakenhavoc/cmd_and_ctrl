package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taste of Blood — Sorcery {B} (Magic 2012, #5573):
//
//	"Taste of Blood deals 1 damage to target player or planeswalker
//	 and you gain 1 life."
//
// Vicious Hunger's body at 1 against Lava Axe's "target player or
// planeswalker" clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e5026da-4859-4f30-824c-a49ac308e91a",
		Name:         "Taste of Blood",
		Completeness: CompletenessFull,
		Targets:      targetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 1)), LifeGain: 1, Answers: game.AnswerRemove},
		OnResolve:    ResolveSteps(damageTheClauseTarget(0, 1), GainLife{Amount: 1}),
	})
}
