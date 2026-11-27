package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Counterbore — Instant {3}{U}{U} (Shadowmoor, #5591):
//
//	"Counter target spell. Search its controller's graveyard, hand, and
//	 library for all cards with the same name as that spell and exile them.
//	 Then that player shuffles."
//
// Quash's body on any spell: see OldB540CounterThenExtract.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1198ec53-9050-417d-812f-692e1378d346",
		Name:         "Counterbore",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.OldB540CounterThenExtract("Counterbore"),
	})
}
