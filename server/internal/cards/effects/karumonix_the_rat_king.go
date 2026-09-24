package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karumonix, the Rat King — Legendary Creature — Phyrexian Rat
// {1}{B}{B}, 3/3:
//
//	"Toxic 1
//	 Other Rats you control have toxic 1.
//	 When Karumonix enters, look at the top five cards of your library.
//	 You may reveal any number of Rat cards from among them and put the
//	 revealed cards into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// The card that proves toxic is CUMULATIVE (CR 702.164b, ADR 0056
// Decision 1): a Rat that prints toxic 1 has a total of 2 under
// Karumonix, because the grant appends through
// game.AppendKeywordAbility, which keeps every toxic instance, and the
// damage tail reads the SUM through game.ToxicTotal.
//
// "Rat cards" is a subtype test on the card in the library, so a
// changeling card is a Rat card too. The enter trigger is Horn of the
// Mark's sentence with "any number" in place of "a".
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "c017f54c-e4c0-411e-b8c1-eb20b1b86c56",
		Name:            "Karumonix, the Rat King",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Rat"}, Others: true, YoursOnly: true}, "toxic 1"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Karumonix, the Rat King — look at the top five cards for Rats",
				LookAtTopThenMayTakeToHand(5, OfCreatureType("Rat"), 0,
					"Karumonix, the Rat King — reveal any number of Rat cards and put them into your hand")),
		},
	})
}
