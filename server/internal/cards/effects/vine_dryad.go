package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vine Dryad — Creature — Dryad {3}{G}, 1/3 (Mercadian Masques, #5619):
//
//	"You may exile a green card from your hand rather than pay this
//	 spell's mana cost.
//	 Flash
//	 Forestwalk (This creature can't be blocked as long as defending
//	 player controls a Forest.)"
//
// Cave-In's pitch with a green card (Pitch, no life), on a creature
// spell: the card is exiled as a cost with the Dryad already on the
// stack (CR 601.2h). Flash and forestwalk are printed keywords; flash
// lets it be cast either way at instant speed (CR 702.8a), and
// forestwalk is enforced in block legality (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "990cd7de-66a2-4def-bf00-390b2682e995",
		Name:            "Vine Dryad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "forestwalk"},
		AlternativeCosts: []game.AlternativeCost{
			Pitch("Exile a green card from your hand", 0,
				CardInYourHand("a green card from your hand", OfColor("G")),
				"a green card from your hand"),
		},
	})
}
