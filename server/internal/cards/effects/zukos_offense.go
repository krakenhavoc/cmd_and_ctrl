package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zuko's Offense — Sorcery {R} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Zuko's Offense deals 2 damage to any target."
//
// Shock at sorcery speed: 2 damage to a creature, a planeswalker, a
// battle or a player, skipped if the target is no longer legal as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90a2ec9c-2f1c-4ed4-820d-dfdb6b58c0db",
		Name:         "Zuko's Offense",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(2),
	})
}
