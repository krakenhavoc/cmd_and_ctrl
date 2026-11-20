package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teyo's Lightshield — Creature — Illusion {2}{W}, 0/3 (War of the Spark, #5520):
//
//	"When this creature enters, put a +1/+1 counter on target creature you control."
//
// Kujar Seedsculptor's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1305d49-e543-49e9-9b36-e56f93f63bc3",
		Name:         "Teyo's Lightshield",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Teyo's Lightshield — put a +1/+1 counter on target creature you control",
				cmdr2S46PlusOneCounterOnClauseTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
