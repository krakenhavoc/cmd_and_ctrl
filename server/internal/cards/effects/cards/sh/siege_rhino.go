package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siege Rhino — Creature — Rhino {1}{W}{B}{G}, 4/5 (Khans of Tarkir,
// #5554):
//
//	"Trample
//	 When this creature enters, each opponent loses 3 life and you gain
//	 3 life."
//
// Debris Beetle's enters trigger, which uses the stack (CR 603.6a);
// trample is printed data (CR 702.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1087ce2-e3c2-4ab1-bd08-1dc9ab70f12f",
		Name:            "Siege Rhino",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Siege Rhino — each opponent loses 3 life and you gain 3 life",
				effects.Do(effects.EachOpponentLosesLife{N: 3}, effects.GainLife{Amount: 3})),
		},
	})
}
