package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle Cry — Instant {2}{W} (Ice Age, S82 OLD, #5636):
//
//	"Untap all white creatures you control.
//	 Whenever a creature blocks this turn, it gets +0/+1 until end of turn."
//
// The untap is over the white creatures you control as it resolves; the second
// sentence is a repeating delayed trigger on every block this turn (CR 603.7b),
// once per blocker (CR 509.3a), whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b92afdba-5aea-42c8-bdd8-6ba5ff6cba4f",
		Name:         "Battle Cry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve:    effects.OldB835BattleCry,
	})
}
