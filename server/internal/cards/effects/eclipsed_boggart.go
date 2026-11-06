package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eclipsed Boggart — {B/R}{B/R}{B/R} Creature — Goblin Scout 2/3 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal a Goblin, Swamp, or Mountain card from among them and
//	 put it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Bucky Barnes' look-and-take (LookAtTopThenMayTakeToHand) with a
// three-subtype filter: one creature type and two basic land types, each an
// effective subtype, so a changeling card is a Goblin card and any land with
// the land type counts. Only the controller sees the four; the card taken is
// revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cc3b91b-ad70-4b01-b718-fda3819f84d0",
		Name:         "Eclipsed Boggart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eclipsed Boggart — look at the top four cards; you may reveal a Goblin, Swamp, or Mountain card and put it into your hand",
				LookAtTopThenMayTakeToHand(4, AnySubtype("Goblin", "Swamp", "Mountain"), 1,
					"Eclipsed Boggart — reveal a Goblin, Swamp, or Mountain card and put it into your hand")),
		},
	})
}
