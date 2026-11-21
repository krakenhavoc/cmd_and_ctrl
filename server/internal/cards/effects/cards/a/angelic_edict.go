package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Edict — Sorcery {4}{W} (Foundations, #4465):
//
//	"Exile target creature or enchantment."
//
// Mortify's target clause with Fall to Earth's exile of the still-legal
// target (exileTheFirstTargetIfLegal).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36de8401-ee4e-413f-91ad-06924e39c857",
		Name:         "Angelic Edict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.ExileTheFirstTargetIfLegal)),
	})
}
