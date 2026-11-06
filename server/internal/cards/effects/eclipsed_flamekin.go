package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eclipsed Flamekin — {1}{U/R}{U/R} Creature — Elemental Scout 1/4 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal an Elemental, Island, or Mountain card from among them
//	 and put it into your hand. Put the rest on the bottom of your library
//	 in a random order."
//
// Eclipsed Boggart's enters trigger with Elemental, Island and Mountain:
// each an effective subtype, so a changeling card is an Elemental card and
// any land with the land type counts. Only the controller sees the four; the
// card taken is revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48c641df-d511-4e35-b623-9d1fdb5240da",
		Name:         "Eclipsed Flamekin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eclipsed Flamekin — look at the top four cards; you may reveal an Elemental, Island, or Mountain card and put it into your hand",
				LookAtTopThenMayTakeToHand(4, AnySubtype("Elemental", "Island", "Mountain"), 1,
					"Eclipsed Flamekin — reveal an Elemental, Island, or Mountain card and put it into your hand")),
		},
	})
}
