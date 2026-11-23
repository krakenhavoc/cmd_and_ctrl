package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xindi Surveyors — Creature — Xindi Scientist {2}{G}, 3/2 (Star Trek,
// #4973):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1 counter
//	 on this creature."
//
// Undergrowth Champion's landfall counter without its prevention: one
// trigger per land (CR 603.2), and a creature that has left gets nothing
// (CounterOnThis, CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fdff691-686d-4ce0-bd12-eea676a648b4",
		Name:         "Xindi Surveyors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Xindi Surveyors — put a +1/+1 counter on it (landfall)", game.LandfallPayoff{Counters: 1},
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
