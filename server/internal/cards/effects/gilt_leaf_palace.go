package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gilt-Leaf Palace — Land (Lorwyn, #5593):
//
//	"As this land enters, you may reveal an Elf card from your hand. If you
//	 don't, this land enters tapped.
//	 {T}: Add {B} or {G}."
//
// Secluded Glen's reveal-land entry over Elf cards: the reveal is a
// choice the player may decline (CR 614.12), the card stays in hand
// (CR 701.20b), and the test is the type line of a card in hand, a
// Kindred card or a changeling card counting as an Elf.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85573a3d-2993-491a-8f8d-bbdb844fa84e",
		Name:         "Gilt-Leaf Palace",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersTappedUnlessYouRevealFromHand("Gilt-Leaf Palace", "an Elf card", lrwB1CardOfType("Elf")),
		},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "G")},
	})
}
