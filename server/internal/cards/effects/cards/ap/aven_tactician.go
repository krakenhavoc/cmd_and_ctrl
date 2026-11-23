package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Tactician — Creature — Bird Soldier {4}{W}, 2/3 (Dragons of Tarkir,
// #5551):
//
//	"Flying
//	 When this creature enters, bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)"
//
// Bolster is effects.Bolster (CR 701.39): the creature with the least
// toughness is chosen as the trigger resolves, and does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f2d5949b-3cdc-4efd-9841-7d450e12dbbe",
		Name:            "Aven Tactician",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aven Tactician — bolster 1", effects.Do(effects.Bolster{N: 1})),
		},
	})
}
