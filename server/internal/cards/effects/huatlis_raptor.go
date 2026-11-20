package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Huatli's Raptor — Creature — Dinosaur {G}{W}, 2/3 (War of the Spark, #5520):
//
//	"Vigilance
//	 When this creature enters, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
//
// Vigilance is printed data; glistening Sphere's enters proliferate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "999ece12-0f1f-4b20-a600-b7f1a584bbb4",
		Name:            "Huatli's Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Huatli's Raptor — proliferate", Do(Proliferate{})),
		},
	})
}
