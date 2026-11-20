package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ad3564f5-11f7-456b-8881-4db00a18a2ed",
		Name:            "Oblivion Strike",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Targets:         TargetCreature("target creature"),
		OnResolve:       ResolveSteps(Instruction(exileTheFirstTargetIfLegal)),
	})
}
