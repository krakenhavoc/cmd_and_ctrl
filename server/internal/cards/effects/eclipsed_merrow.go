package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eclipsed Merrow — Creature — Merfolk Scout {W/U}{W/U}{W/U}, 2/3 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal a Merfolk, Plains, or Island card from among them and
//	 put it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Ingenious Smith's entry trigger with a different filter: the
// LookAtTopThenMayTakeToHand sentence over AnySubtype, so a changeling card
// is a Merfolk card and a nonbasic land with the Plains or Island type
// counts. Only the card taken is revealed; the rest go to the bottom in a
// random order.
//
// No simplification.
func init() {
	const lookLabel = "Eclipsed Merrow — look at the top four cards; you may reveal a Merfolk, Plains, or Island card and put it into your hand"
	Register(Spec{
		OracleID:     "a3a00b62-f07e-44ef-8d97-22a3e7a06f72",
		Name:         "Eclipsed Merrow",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(lookLabel,
				LookAtTopThenMayTakeToHand(4, AnySubtype("Merfolk", "Plains", "Island"), 1,
					"Eclipsed Merrow — reveal a Merfolk, Plains, or Island card and put it into your hand")),
		},
	})
}
