package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Edict — Sorcery {4}{W} (Foundations, #4465):
//
//	"Exile target creature or enchantment."
//
// Mortify's target clause with Fall to Earth's exile of the still-legal
// target (exileTheFirstTargetIfLegal).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36de8401-ee4e-413f-91ad-06924e39c857",
		Name:         "Angelic Edict",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
		OnResolve:    ResolveSteps(Instruction(exileTheFirstTargetIfLegal)),
	})
}
