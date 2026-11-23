package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verdant Automaton — Artifact Creature — Construct {2}, 1/2 (Aether Revolt,
// #5536):
//
//	"{3}{G}: Put a +1/+1 counter on this creature."
//
// Jungle Delver's counter row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3af1c69-42ae-4d29-a9f1-72f862d26303",
		Name:         "Verdant Automaton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{G}"),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
