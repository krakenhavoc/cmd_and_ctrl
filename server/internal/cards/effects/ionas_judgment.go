package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iona's Judgment — Sorcery {4}{W} (Worldwake, #5582):
//
//	"Exile target creature or enchantment."
//
// Angelic Edict's text and body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abbfbe50-4caf-48d0-b8f1-f2c9212bcd2f",
		Name:         "Iona's Judgment",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
		OnResolve:    ResolveSteps(Instruction(exileTheFirstTargetIfLegal)),
	})
}
