package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lava Axe — Sorcery {4}{R} (Portal, #5630):
//
//	"Lava Axe deals 5 damage to target player or planeswalker."
//
// Boros Charm's "target player or planeswalker" clause
// (targetPlayerOrPlaneswalker) with the shared damage body
// (damageToFirstTarget). Damage to a player is life loss (CR 120.3a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "387b6b07-a283-412d-94c3-f7f1dc76e858",
		Name:         "Lava Axe",
		Completeness: CompletenessFull,
		Targets:      targetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(5),
	})
}
