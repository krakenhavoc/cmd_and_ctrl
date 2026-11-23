package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloom Hulk — Creature — Plant Elemental {3}{G}, 4/4 (War of the Spark, #5520):
//
//	"When this creature enters, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
//
// Glistening Sphere's enters proliferate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c99896a8-8b58-4946-8f2d-b4e2562f278a",
		Name:         "Bloom Hulk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bloom Hulk — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
