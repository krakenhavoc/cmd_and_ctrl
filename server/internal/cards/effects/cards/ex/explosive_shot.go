package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explosive Shot — Sorcery {1}{R} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Explosive Shot deals 4 damage to target creature."
//
// Flame Slash's body: 4 damage to one creature, whoever controls it,
// skipped if the target is no longer legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6081d17-599b-4fd2-bd68-55eca939213a",
		Name:         "Explosive Shot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.DamageToTarget(0, 4)),
		},
		OnResolve: effects.DamageToFirstTargetStep(4),
	})
}
