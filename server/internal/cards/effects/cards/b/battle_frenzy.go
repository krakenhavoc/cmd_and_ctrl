package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle Frenzy — Instant {2}{R} (Ice Age, S82 OLD, #5636):
//
//	"Green creatures you control get +1/+1 until end of turn.
//	 Nongreen creatures you control get +1/+0 until end of turn."
//
// Two mass pumps over the creatures you control as it resolves (CR 611.2c):
// each creature gets exactly one of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fca46ec-ec1a-4e6c-bb0e-cdb834b38c92",
		Name:         "Battle Frenzy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve:    effects.OldB834BattleFrenzy,
	})
}
