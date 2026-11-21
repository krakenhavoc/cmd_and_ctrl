package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ninth Bridge Patrol — Creature — Dwarf Soldier {1}{W}, 1/1 (Jumpstart 2022):
//
//	"Whenever another creature you control leaves the battlefield, put a
//	 +1/+1 counter on this creature."
//
// Lunarch Veteran's condition (another creature, under the source's
// controller as it last existed, any destination) with a counter on the
// source. A Patrol that left with it gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21ad3d41-54e9-48f9-b150-be65f52ae464",
		Name:         "Ninth Bridge Patrol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AnotherCreatureYouControlLeft, "Ninth Bridge Patrol — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
