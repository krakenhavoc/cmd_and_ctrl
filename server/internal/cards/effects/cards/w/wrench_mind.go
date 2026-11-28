package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrench Mind — Sorcery {B}{B} (Mirrodin, #5606):
//
//	"Target player discards two cards unless they discard an artifact card."
//
// Teferi Akosa's "two cards unless one of a kind" discard, asked of the
// targeted player: they pick the cards (CR 701.9a), and a pick pays the clause
// when it is two cards or one artifact card. A hand of one card is asked the
// one discard, whatever it is; an empty hand is asked nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "484eeb13-fe74-45fe-b089-7beec87cfcdb",
		Name:         "Wrench Mind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.OldB607WrenchMindDiscard,
	})
}
