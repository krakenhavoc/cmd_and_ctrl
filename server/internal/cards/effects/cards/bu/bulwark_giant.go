package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bulwark Giant — Creature — Giant Soldier {5}{W}, 3/6 (War of the Spark, #5520):
//
//	"When this creature enters, you gain 5 life."
//
// Thragtusk's enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f606bd15-b380-447f-b147-72f811266593",
		Name:         "Bulwark Giant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 5},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bulwark Giant — you gain 5 life", effects.Do(effects.GainLife{Amount: 5})),
		},
	})
}
