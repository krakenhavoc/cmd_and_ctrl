package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grip of Amnesia — Instant {1}{U} (Judgment, #5611):
//
//	"Counter target spell unless its controller exiles all cards from
//	 their graveyard.
//	 Draw a card."
//
// The spell's controller is asked, through the same counter-unless prompt
// ward uses, and pays by exiling every card of their graveyard; the table
// waits for the answer. The draw follows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0144d9bb-d3cb-4c23-b133-6653d7c3a76d",
		Name:         "Grip of Amnesia",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.ResolveSteps(effects.OldB621CounterUnlessItExilesItsGraveyard, effects.DrawCards{N: 1}),
	})
}
