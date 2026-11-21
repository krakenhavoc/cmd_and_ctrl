package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderwolf Cavalry — Creature — Astartes Warrior {4}{W}, 4/4 (Warhammer
// 40,000 Commander, S82 GAP batch 196, #5948):
//
//	"First strike
//	 Crushing Teeth — Whenever this creature deals combat damage to a player, put
//	 a +1/+1 counter on each other creature you control."
//
// Mikaeus the Lunarch's counter on each other creature: the set is the other
// creatures its controller controls as the trigger resolves, snapshotted before
// the first counter lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b36df26-6a41-4714-95a8-fd02e2fc845f",
		Name:            "Thunderwolf Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Thunderwolf Cavalry — a +1/+1 counter on each other creature you control",
				effects.PutCounterOnEachOtherCreatureYouControl),
		},
	})
}
