package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rot Shambler — Creature — Fungus {1}{G}, 1/1 (Battle for Zendikar,
// #5547):
//
//	"Whenever another creature you control dies, put a +1/+1 counter
//	 on this creature."
//
// Unruly Mob's trigger: "you control" looks back to the creature as it
// died (CR 603.10a), so an opponent's creature dying gives nothing, and
// each death is its own trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88c187fd-b3d3-4218-8e63-380a9ca9f1b4",
		Name:         "Rot Shambler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AllOf(effects.AnotherCreatureDied, effects.ACreatureYouControlDied),
				"Rot Shambler — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
