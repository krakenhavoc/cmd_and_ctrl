package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fblthp, Lost on the Range — Legendary Creature — Homunculus {1}{U}{U}, 1/1:
//
//	"Ward {2}
//	 You may look at the top card of your library any time.
//	 The top card of your library has plot. The plot cost is equal to
//	 its mana cost.
//	 You may plot nonland cards from the top of your library."
//
// The first permanent that gives a special action to ANOTHER card
// (#1391, ADR 0062 amendment 2026-09-24). Each printed line maps to
// one declaration:
//
//   - Ward {2} is the shared ward trigger.
//   - "Look at the top card any time" is game.LibraryTopOwner. The top
//     card is known to Fblthp's controller alone, the same as
//     Realmwalker and Bolas's Citadel. Opponents see it only if
//     something else reveals it (Courser of Kruphix).
//   - The two plot sentences are one grant, PlotFromTopOfLibrary. The
//     nonland top card may be plotted for its mana cost, or for its
//     own plot cost if it prints plot. That is the ordinary plot
//     special action (sorcery timing, face-up exile, a free cast on a
//     later turn) taken from the library instead of the hand.
//
// Scryfall lists Plot among the card's keywords. Fblthp itself does
// NOT have plot: it gives plot to the top card of your library. So
// there is no Plot(...) in SpecialActions, and a Fblthp in hand cannot
// be plotted, which is the printed card.
//
// Edge cases, each the rules' answer and not a simplification:
//   - a top card with no mana cost cannot be plotted for its mana cost
//     (CR 118.6: a cost based on the mana cost of an object with no
//     mana cost is unpayable);
//   - {X} counts as 0, and the plotted cast is free anyway;
//   - a split card costs both halves combined (CR 709.4b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "764f6412-c4ff-4eec-9a9d-870661f97f8b",
		Name:         "Fblthp, Lost on the Range",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Fblthp, Lost on the Range — ward {2}"),
		},
		LibraryTopVisible:   game.LibraryTopOwner,
		SpecialActionGrants: []game.SpecialActionGrant{PlotFromTopOfLibrary()},
	})
}
