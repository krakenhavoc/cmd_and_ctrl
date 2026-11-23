package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Makeshift Battalion — Creature — Human Soldier {2}{W}, 3/2 (Ravnica
// Remastered, #5948):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, put a +1/+1 counter on this creature."
//
// Battalion is Boros Strike-Captain's group test (gtcBattalion); the counter
// goes on this creature only while it is still the permanent the ability came
// from (CR 400.7), through the counter window.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b6e2feb-9e8f-4912-ada0-1614502d9813",
		Name:         "Makeshift Battalion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GtcBattalion("Makeshift Battalion — put a +1/+1 counter on it (battalion)",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
