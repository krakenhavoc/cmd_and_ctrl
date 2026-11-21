package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shredding Winds — Instant {2}{G} (Theros, #5561):
//
//	"Shredding Winds deals 7 damage to target creature with flying."
//
// Seven damage to a creature with flying, read as it is cast and again
// as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f914186-082e-4a85-a1cc-e9788074637b",
		Name:         "Shredding Winds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 7)), Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve:    effects.DamageToFirstTargetStep(7),
	})
}
