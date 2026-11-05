package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gin'tak Charge — Sorcery {R} (Star Trek Commander, #4974):
//
//	"Gin'tak Charge deals 2 damage to any target."
//
// Shock's damage at sorcery speed (the timing is the card type's), through
// the shared damageToFirstTarget body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "286213a8-ae53-4f97-9f81-e13bed462d0a",
		Name:         "Gin'tak Charge",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove}, // #4734: damages a creature, planeswalker or player
		OnResolve:    damageToFirstTarget(2),
	})
}
