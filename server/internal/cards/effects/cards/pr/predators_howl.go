package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Predator's Howl — Instant {3}{G} (Jumpstart 2022, #5948):
//
//	"Create a 2/2 green Wolf creature token.
//	 Morbid — Create three 2/2 green Wolf creature tokens instead if a
//	 creature died this turn."
//
// Hunger of the Howlpack's morbid read (any creature, any controller) as the
// spell resolves, choosing between one Wolf and three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1180796-6e26-4b45-9263-130620d702e5",
		Name:         "Predator's Howl",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
		OnResolve:    effects.GapB171Wolves,
	})
}
