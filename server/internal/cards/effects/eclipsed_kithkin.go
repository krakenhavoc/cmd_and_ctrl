package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eclipsed Kithkin — {G/W}{G/W} Creature — Kithkin Scout 2/1 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal a Kithkin, Forest, or Plains card from among them and
//	 put it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Eclipsed Boggart's enters trigger with Kithkin, Forest and Plains: each
// an effective subtype, so a changeling card is a Kithkin card and any land
// with the land type counts. Only the controller sees the four; the card
// taken is revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6227840-6e3e-48c7-b4a4-42cfd8f14b68",
		Name:         "Eclipsed Kithkin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eclipsed Kithkin — look at the top four cards; you may reveal a Kithkin, Forest, or Plains card and put it into your hand",
				LookAtTopThenMayTakeToHand(4, AnySubtype("Kithkin", "Forest", "Plains"), 1,
					"Eclipsed Kithkin — reveal a Kithkin, Forest, or Plains card and put it into your hand")),
		},
	})
}
