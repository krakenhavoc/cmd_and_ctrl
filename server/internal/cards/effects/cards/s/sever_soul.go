package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sever Soul — Sorcery {3}{B}{B} (Eighth Edition, #5607):
//
//	"Destroy target nonblack creature. It can't be regenerated. You gain
//	 life equal to its toughness."
//
// The toughness is read as the spell resolves, before the destruction (CR
// 608.2h); the life is gained even if the creature survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "577d027a-96c3-46fe-880b-f8b5dd3f3a3d",
		Name:         "Sever Soul",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve:    effects.SeverSoulResolve,
	})
}
