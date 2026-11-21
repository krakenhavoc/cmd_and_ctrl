package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bee Sting — Sorcery {3}{G} (Portal, #5630):
//
//	"Bee Sting deals 2 damage to any target."
//
// Shock's body as a sorcery (damageToFirstTarget). A target that
// became illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f637d525-2f29-488a-9269-8e5aa377fbb7",
		Name:         "Bee Sting",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
