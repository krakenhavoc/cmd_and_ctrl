package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huatli's Raptor — Creature — Dinosaur {G}{W}, 2/3 (War of the Spark, #5520):
//
//	"Vigilance
//	 When this creature enters, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
//
// Vigilance is printed data; glistening Sphere's enters proliferate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "999ece12-0f1f-4b20-a600-b7f1a584bbb4",
		Name:            "Huatli's Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Huatli's Raptor — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
