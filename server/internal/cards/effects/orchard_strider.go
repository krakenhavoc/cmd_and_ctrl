package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c6906032-e08b-478f-bc71-cb1184f60bad",
		Name:         "Orchard Strider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Orchard Strider — create two Food tokens",
				Do(CreateToken{Template: FoodToken(), N: 2})),
		},
		Activated: []ActivatedAbility{BasicLandcycling("{1}{G}")},
	})
}
