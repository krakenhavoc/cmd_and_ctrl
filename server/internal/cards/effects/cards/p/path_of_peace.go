package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Path of Peace — {3}{W} Sorcery (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Destroy target creature. Its owner gains 4 life."
//
// The owner, not the controller, gains the life; it is read before the
// destruction (S82OldB702PathOfPeace).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7593cf8-4dcb-473b-a2ef-180fffe66738",
		Name:         "Path of Peace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.S82OldB702PathOfPeace,
	})
}
