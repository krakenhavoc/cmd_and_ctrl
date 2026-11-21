package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heirs of Stromkirk — {2}{R}{R} Creature — Vampire 2/2 (Avacyn Restored, #5570):
//
//	"Intimidate (This creature can't be blocked except by artifact
//	 creatures and/or creatures that share a color with it.)
//	 Whenever this creature deals combat damage to a player, put a
//	 +1/+1 counter on it."
//
// Intimidate (CR 702.13) and Markov Blademaster's combat-damage
// counter, written with the CounterOnThis primitive.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7149de06-d676-47bc-a50f-e1058d89c890",
		Name:            "Heirs of Stromkirk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Heirs of Stromkirk — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
