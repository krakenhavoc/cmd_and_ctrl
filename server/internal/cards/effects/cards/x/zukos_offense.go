package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "90a2ec9c-2f1c-4ed4-820d-dfdb6b58c0db",
		Name:         "Zuko's Offense",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
