package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chase Stein, Runaway — Legendary Creature — Human Hero {1}{R}, 2/2
// (Marvel Super Heroes Commander):
//
//	"{T}, Discard a card: Exile the top card of your library. Until the
//	 end of your next turn, you may play that card."
//
// Cori Mountain Monastery's impulse body (b20ExileTopUntilEndOfNextTurn,
// "play", so a land is not stranded) with a discard in the cost. The
// discard is chosen at announcement (CR 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "161ff063-b784-43c7-b608-f78be3daa5b2",
		Name:         "Chase Stein, Runaway",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a card: Exile the top card of your library. Until the end of your next turn, you may play that card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
			Effect:  effects.B20ExileTopUntilEndOfNextTurn,
		}},
	})
}
