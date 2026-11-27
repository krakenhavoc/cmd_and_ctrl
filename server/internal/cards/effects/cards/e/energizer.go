package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Energizer — Artifact Creature — Juggernaut {4}, 2/2 (Tempest, S82 OLD
// batch 760, #5628):
//
//	"{2}, {T}: Put a +1/+1 counter on this creature."
//
// A mana-and-tap activation whose counter goes on the Energizer only while
// it is still the permanent that activated (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "708b0321-7de9-4ffd-9d7a-9f4813dd542e",
		Name:         "Energizer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}
