package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blessed Light — Instant {4}{W} (Dominaria, #5530):
//
//	"Exile target creature or enchantment."
//
// Angelic Edict's sentence: the target, if it is still legal as the spell
// resolves (CR 608.2b), is exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "918b3860-e96c-45b5-b0e6-e82cad9f304d",
		Name:         "Blessed Light",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
		OnResolve:    ResolveSteps(Instruction(exileTheFirstTargetIfLegal)),
	})
}
