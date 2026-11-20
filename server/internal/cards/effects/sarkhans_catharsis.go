package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarkhan's Catharsis — Instant {4}{R} (War of the Spark, #5520):
//
//	"Sarkhan's Catharsis deals 5 damage to target player or planeswalker."
//
// Lava Axe's body, at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "787935ab-26a6-4b4a-9c14-e106160e951d",
		Name:         "Sarkhan's Catharsis",
		Completeness: CompletenessFull,
		Targets:      targetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(5),
	})
}
