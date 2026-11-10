package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wickerfolk Indomitable — Artifact Creature — Scarecrow {3}{B}, 4/3
// (Aetherdrift, #4464):
//
//	"You may cast this card from your graveyard by paying 2 life and
//	 sacrificing an artifact or creature in addition to paying its other
//	 costs."
//
// Not an alternative cost (CR 118.8): the {3}{B} is still paid, and the
// life and the sacrifice are added to it. The offer is bound to the
// graveyard (retrace's model), so a hand cast pays only the mana cost and
// a graveyard cast must pay all three. The sacrifice is paid with the
// spell already on the stack, so the Scarecrow itself can't be it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "730d02c7-2f8a-47d3-99f0-e4ada7f6255c",
		Name:          "Wickerfolk Indomitable",
		Completeness:  CompletenessFull,
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			b13CastFromGraveyardPayingLifeAndSacrificing("{3}{B}", 2, "an artifact or creature", Or(Artifact(), Creature())),
		},
	})
}
