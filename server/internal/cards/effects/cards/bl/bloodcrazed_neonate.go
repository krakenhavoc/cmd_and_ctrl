package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodcrazed Neonate — Creature — Vampire {1}{R}, 2/1 (Innistrad, S82 older
// sets batch 449, #5572):
//
//	"This creature attacks each combat if able.
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it."
//
// Planet Killer's requirement and Havengul Vampire's counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7319eec8-48ba-4238-bce1-450abcae73a6",
		Name:         "Bloodcrazed Neonate",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Bloodcrazed Neonate — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
