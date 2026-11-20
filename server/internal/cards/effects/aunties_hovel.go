package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auntie's Hovel — Land (Lorwyn, #5593):
//
//	"As this land enters, you may reveal a Goblin card from your hand. If you
//	 don't, this land enters tapped.
//	 {T}: Add {B} or {R}."
//
// Secluded Glen's reveal-land entry over Goblin cards: the reveal is a
// choice the player may decline (CR 614.12), the card stays in hand
// (CR 701.20b), and the test is the type line of a card in hand, a
// Kindred card or a changeling card counting as a Goblin.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "245469ff-72b6-4846-8a82-a1d29f4d09bb",
		Name:         "Auntie's Hovel",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersTappedUnlessYouRevealFromHand("Auntie's Hovel", "a Goblin card", lrwB1CardOfType("Goblin")),
		},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "R")},
	})
}
