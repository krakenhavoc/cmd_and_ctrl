package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiora's Dambreaker — Creature — Leviathan {5}{U}, 5/6 (War of the Spark, #5520):
//
//	"When this creature enters, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
//
// Glistening Sphere's enters proliferate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "795b903b-5260-4c06-b26f-a89cecdfa1f8",
		Name:         "Kiora's Dambreaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kiora's Dambreaker — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
