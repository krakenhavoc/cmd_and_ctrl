package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bulwark Giant — Creature — Giant Soldier {5}{W}, 3/6 (War of the Spark, #5520):
//
//	"When this creature enters, you gain 5 life."
//
// Thragtusk's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f606bd15-b380-447f-b147-72f811266593",
		Name:         "Bulwark Giant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 5},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bulwark Giant — you gain 5 life", Do(GainLife{Amount: 5})),
		},
	})
}
