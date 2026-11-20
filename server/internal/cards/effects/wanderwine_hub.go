package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wanderwine Hub — Land (Lorwyn, #5593):
//
//	"As this land enters, you may reveal a Merfolk card from your hand. If you
//	 don't, this land enters tapped.
//	 {T}: Add {W} or {U}."
//
// Secluded Glen's reveal-land entry over Merfolk cards: the reveal is a
// choice the player may decline (CR 614.12), the card stays in hand
// (CR 701.20b), and the test is the type line of a card in hand, a
// Kindred card or a changeling card counting as a Merfolk.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3b46bd6-b3ef-452d-a916-995c44f1da07",
		Name:         "Wanderwine Hub",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersTappedUnlessYouRevealFromHand("Wanderwine Hub", "a Merfolk card", lrwB1CardOfType("Merfolk")),
		},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "U")},
	})
}
