package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ironpaw Aspirant — Creature — Cat Warrior {1}{W}, 1/2 (The Lost Caverns
// of Ixalan, #4470):
//
//	"When this creature enters, put a +1/+1 counter on target creature."
//
// Cloudbound Moogle's enters trigger: the target is any creature, the
// Aspirant itself included, and a target that has gone by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fab5f694-4a9c-4e72-b164-f236e8212ff7",
		Name:         "Ironpaw Aspirant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ironpaw Aspirant — put a +1/+1 counter on target creature", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
