package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oblivion Strike — Sorcery {3}{B} (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 Exile target creature."
//
// Devoid makes it colourless (CR 702.114a). Angelic Edict's body: the
// target is exiled if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad3564f5-11f7-456b-8881-4db00a18a2ed",
		Name:            "Oblivion Strike",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Targets:         effects.TargetCreature("target creature"),
		OnResolve:       effects.ResolveSteps(effects.Instruction(effects.ExileTheFirstTargetIfLegal)),
	})
}
