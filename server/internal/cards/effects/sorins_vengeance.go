package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sorin's Vengeance — Sorcery {4}{B}{B}{B} (Magic 2012, #5573):
//
//	"Sorin's Vengeance deals 10 damage to target player or
//	 planeswalker and you gain 10 life."
//
// Vicious Hunger's body at 10 against Lava Axe's "target player or
// planeswalker" clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75d9c036-4f0d-4b55-b0a4-096ca84748ca",
		Name:         "Sorin's Vengeance",
		Completeness: CompletenessFull,
		Targets:      targetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 10)), LifeGain: 10, Answers: game.AnswerRemove},
		OnResolve:    ResolveSteps(damageTheClauseTarget(0, 10), GainLife{Amount: 10}),
	})
}
