package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orchard Strider — Creature — Treefolk {4}{G}{G}, 6/4 (Foundations Jumpstart):
//
//	"When this creature enters, create two Food tokens.
//	 Basic landcycling {1}{G}"
//
// The two Foods are one creation (CR 111.2), as Bristlebud Farmer's;
// Basic landcycling is the shared Typecycling ability (CR 702.29).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6906032-e08b-478f-bc71-cb1184f60bad",
		Name:         "Orchard Strider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Orchard Strider — create two Food tokens",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 2})),
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{1}{G}")},
	})
}
