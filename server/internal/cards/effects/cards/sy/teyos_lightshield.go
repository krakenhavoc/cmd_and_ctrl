package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teyo's Lightshield — Creature — Illusion {2}{W}, 0/3 (War of the Spark, #5520):
//
//	"When this creature enters, put a +1/+1 counter on target creature you control."
//
// Kujar Seedsculptor's enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1305d49-e543-49e9-9b36-e56f93f63bc3",
		Name:         "Teyo's Lightshield",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Teyo's Lightshield — put a +1/+1 counter on target creature you control",
				effects.Cmdr2S46PlusOneCounterOnClauseTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
