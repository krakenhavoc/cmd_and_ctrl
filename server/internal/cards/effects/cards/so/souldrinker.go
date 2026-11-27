package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Souldrinker — Creature — Spirit {3}{B}, 2/2 (Tempest, S82 OLD batch
// 759, #5628):
//
//	"Pay 3 life: Put a +1/+1 counter on this creature."
//
// A life payment is a cost (CR 118.4), refused by a player who has fewer
// than 3 life (CR 119.4) and paid at announcement; the counter goes on the
// Souldrinker only while it is still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "624d4bda-448f-43d1-92b6-53a77afe2374",
		Name:         "Souldrinker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 3 life: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.PayLife(3),
			Effect:  effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}
