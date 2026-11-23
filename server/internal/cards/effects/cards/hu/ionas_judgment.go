package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iona's Judgment — Sorcery {4}{W} (Worldwake, #5582):
//
//	"Exile target creature or enchantment."
//
// Angelic Edict's text and body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abbfbe50-4caf-48d0-b8f1-f2c9212bcd2f",
		Name:         "Iona's Judgment",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.ExileTheFirstTargetIfLegal)),
	})
}
