package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yotian Dissident — Creature — Human Artificer {G}{W}, 1/1 (The Brothers'
// War, #5482):
//
//	"Whenever an artifact you control enters, put a +1/+1 counter on
//	 target creature you control."
//
// The target is chosen as the trigger goes on the stack and checked again
// as it resolves; a creature that has gone gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ad61622-a70f-423c-99e3-b1ef6ead0b31",
		Name:         "Yotian Dissident",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ArtifactEnteredUnderYourControl,
				"Yotian Dissident — put a +1/+1 counter on target creature you control",
				effects.B36CounterOnChosenAnimal),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
