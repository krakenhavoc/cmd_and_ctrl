package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromoka Captain — Creature — Human Soldier {2}{W}, 1/1 (Dragons of Tarkir, S82 old
// batch 330):
//
//	"First strike
//	 Whenever this creature attacks, bolster 1. (Choose a creature with the
//	 least toughness among creatures you control and put a +1/+1 counter on
//	 it.)"
//
// Printed first strike and an attack trigger that bolsters (CR 701.39a): the
// creature is chosen as the trigger resolves, so it may be the Captain or any
// creature tied for least toughness, and several ties are the controller's pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "74333718-2a05-4643-9a37-013df39b32e0",
		Name:            "Dromoka Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Dromoka Captain — bolster 1", effects.Do(effects.Bolster{N: 1})),
		},
	})
}
