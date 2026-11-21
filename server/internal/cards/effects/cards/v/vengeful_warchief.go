package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vengeful Warchief — {4}{B} Creature — Orc Warrior 4/3:
//
//	"Whenever you lose life for the first time each turn, put a +1/+1
//	 counter on this creature. (Damage causes loss of life.)"
//
// The same trigger as Gonti's Machinations
// (WheneverYouLoseLifeForTheFirstTimeEachTurn, #2540). A Warchief that
// has left the battlefield by the time the trigger resolves puts the
// counter on nothing (thisStillHere).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be3dd4f6-794d-4bce-b71e-4b3477d5d388",
		Name:         "Vengeful Warchief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouLoseLifeForTheFirstTimeEachTurn("Vengeful Warchief — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
