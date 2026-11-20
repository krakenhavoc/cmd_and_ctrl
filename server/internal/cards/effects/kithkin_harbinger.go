package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kithkin Harbinger — Creature — Kithkin Wizard {2}{W}, 1/3 (Lorwyn, #5593):
//
//	"When this creature enters, you may search your library for a Kithkin
//	 card, reveal it, then shuffle and put that card on top."
//
// The Harbinger enters tutor (lrwB1Harbinger): a "you may" asked as the
// trigger resolves (CR 603.5), a search for a Kithkin card that may find
// nothing (CR 701.23b), shown to the table, and placed on top after the
// shuffle. A Kindred card of the type counts, and so does a changeling
// card (CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3d68848-e8cc-497c-951b-e7111f6d5142",
		Name:         "Kithkin Harbinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			lrwB1Harbinger("Kithkin Harbinger", "a Kithkin card", lrwB1CardOfType("Kithkin")),
		},
	})
}
