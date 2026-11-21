package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenhill Flock — Creature — Bird {3}{U} (The Hobbit), 1/2:
//
//	"Flying
//	 Whenever you draw a card, put a +1/+1 counter on this creature."
//
// #4454, S76. Chasm Skulker's draw trigger: one counter per card drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7f961374-00b0-40df-83eb-9ced2bb519aa",
		Name:            "Ravenhill Flock",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Ravenhill Flock — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
